package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"html"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html/charset"
)

const (
	emailFilename  = "email.md"
	attachmentsDir = "attachments"
)

// EmailMetadata holds parsed email information
type EmailMetadata struct {
	Cc         []string
	Date       string
	From       string
	InReplyTo  string
	MessageID  string
	References string
	Subject    string
	To         []string
}

// Attachment represents an extracted attachment
type Attachment struct {
	Filename string
	Path     string
	Size     int64
}

// ExtractionResult contains all extracted email data
type ExtractionResult struct {
	Attachments  []Attachment
	EmailName    string
	Markdown     string
	MarkdownFile string
	Metadata     EmailMetadata
	OutputDir    string
}

func main() {
	var (
		cleanup = flag.Bool("cleanup", false, "Clean up extraction directory after reading")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <eml_file> [output_directory]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Extract content and attachments from .eml files to markdown format.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  eml_file           Path to the .eml file to extract\n")
		fmt.Fprintf(os.Stderr, "  output_directory   Optional: Base directory for extraction\n")
		fmt.Fprintf(os.Stderr, "                     Default: Same directory as .eml file\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s message.eml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s ~/Downloads/email.eml ~/Documents/extracted\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --cleanup message.eml\n\n", os.Args[0])
	}

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	emlPath := args[0]
	var outputDir string
	if len(args) >= 2 {
		outputDir = args[1]
	}

	// Extract email
	result, err := extractEmailContent(emlPath, outputDir, *cleanup)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error extracting email: %v\n", err)
		os.Exit(1)
	}

	// Print summary and content
	printExtractionSummary(result)

	// Cleanup if requested
	if *cleanup {
		cleanupExtraction(result.OutputDir)
	} else {
		fmt.Fprintf(os.Stderr, "\n💡 Tip: Use --cleanup flag to automatically remove extraction directory after reading\n")
		fmt.Fprintf(os.Stderr, "   Or manually clean up: rm -rf \"%s\"\n", result.OutputDir)
	}
}

func extractEmailContent(emlPath, outputDir string, cleanup bool) (*ExtractionResult, error) {
	// Expand and resolve path
	emlPath = expandPath(emlPath)
	if _, err := os.Stat(emlPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("email file not found: %s", emlPath)
	}

	emlDir := filepath.Dir(emlPath)
	emlFilename := strings.TrimSuffix(filepath.Base(emlPath), ".eml")

	fmt.Fprintf(os.Stderr, "📧 Extracting: %s\n", emlPath)

	// Parse email
	f, err := os.Open(emlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open email file: %w", err)
	}
	defer f.Close()

	msg, err := mail.ReadMessage(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse email: %w", err)
	}

	// Get subject for folder name
	subject := decodeHeader(msg.Header.Get("Subject"))
	var folderName string
	if subject != "" {
		folderName = sanitizeFilename(subject)
		if len(folderName) > 100 {
			folderName = folderName[:100]
		}
	} else {
		folderName = emlFilename
	}

	if !strings.HasSuffix(folderName, "_email") {
		folderName += "_email"
	}

	// Determine output directory
	if outputDir == "" {
		outputDir = filepath.Join(emlDir, folderName)
	} else {
		outputDir = expandPath(outputDir)
		if !strings.HasSuffix(outputDir, folderName) {
			outputDir = filepath.Join(outputDir, folderName)
		}
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Fprintf(os.Stderr, "📁 Output to: %s\n", outputDir)

	// Parse metadata
	metadata := parseEmailMetadata(msg)

	// Extract body
	body, err := getEmailBody(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to extract email body: %w", err)
	}

	// Extract attachments
	fmt.Fprintf(os.Stderr, "\n📎 Extracting attachments...\n")
	attachments, err := extractAttachments(msg, outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error extracting attachments: %v\n", err)
		attachments = []Attachment{} // Continue with empty attachments
	}

	// Create markdown content
	markdown := createEmailMarkdown(metadata, body, attachments)

	// Save markdown file
	markdownPath := filepath.Join(outputDir, emailFilename)
	if err := os.WriteFile(markdownPath, []byte(markdown), 0644); err != nil {
		return nil, fmt.Errorf("failed to write markdown file: %w", err)
	}

	return &ExtractionResult{
		Markdown:     markdown,
		Metadata:     metadata,
		Attachments:  attachments,
		OutputDir:    outputDir,
		MarkdownFile: markdownPath,
		EmailName:    folderName,
	}, nil
}

func parseEmailMetadata(msg *mail.Message) EmailMetadata {
	header := msg.Header

	// Parse date
	dateStr := header.Get("Date")
	formattedDate := dateStr
	if t, err := mail.ParseDate(dateStr); err == nil {
		formattedDate = t.Format("2006-01-02 15:04:05")
	}

	// Parse addresses
	fromAddr := formatEmailAddress(header.Get("From"))
	toAddrs := parseAddressList(header.Get("To"))
	ccAddrs := parseAddressList(header.Get("Cc"))

	return EmailMetadata{
		From:       fromAddr,
		To:         toAddrs,
		Cc:         ccAddrs,
		Subject:    decodeHeader(header.Get("Subject")),
		Date:       formattedDate,
		MessageID:  header.Get("Message-ID"),
		InReplyTo:  header.Get("In-Reply-To"),
		References: header.Get("References"),
	}
}

func getEmailBody(msg *mail.Message) (string, error) {
	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		// Simple text email
		body, err := io.ReadAll(msg.Body)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", fmt.Errorf("failed to parse content type: %w", err)
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			return "[No readable content found]", nil
		}

		mr := multipart.NewReader(msg.Body, boundary)
		return extractMultipartBody(mr)
	}

	// Single part message
	return extractPartBody(msg.Body, mediaType, params)
}

func extractMultipartBody(mr *multipart.Reader) (string, error) {
	var plainText, htmlText string

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		partContentType := part.Header.Get("Content-Type")
		partMediaType, partParams, err := mime.ParseMediaType(partContentType)
		if err != nil {
			continue
		}

		// Skip attachments
		if disposition := part.Header.Get("Content-Disposition"); strings.Contains(disposition, "attachment") {
			continue
		}

		// Handle nested multipart
		if strings.HasPrefix(partMediaType, "multipart/") {
			if boundary := partParams["boundary"]; boundary != "" {
				nestedMr := multipart.NewReader(part, boundary)
				if body, err := extractMultipartBody(nestedMr); err == nil && body != "" {
					if strings.Contains(partMediaType, "alternative") && htmlText == "" {
						htmlText = body
					} else if plainText == "" {
						plainText = body
					}
				}
			}
			continue
		}

		// Extract text parts
		body, err := extractPartBody(part, partMediaType, partParams)
		if err != nil || body == "" {
			continue
		}

		if partMediaType == "text/plain" && plainText == "" {
			plainText = body
		} else if partMediaType == "text/html" && htmlText == "" {
			htmlText = body
		}
	}

	// Prefer plain text over HTML
	if plainText != "" {
		return plainText, nil
	}
	if htmlText != "" {
		return htmlToText(htmlText), nil
	}

	return "[No readable content found]", nil
}

func extractPartBody(r io.Reader, mediaType string, params map[string]string) (string, error) {
	// Read content
	content, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	// Decode if needed
	if enc := params["charset"]; enc != "" && enc != "utf-8" && enc != "UTF-8" {
		if decoded, err := decodeContent(content, enc); err == nil {
			content = decoded
		}
	}

	text := string(content)

	// Convert HTML to text if needed
	if mediaType == "text/html" {
		text = htmlToText(text)
	}

	return strings.TrimSpace(text), nil
}

func extractAttachments(msg *mail.Message, outputDir string) ([]Attachment, error) {
	var attachments []Attachment

	contentType := msg.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		return attachments, nil
	}

	boundary := params["boundary"]
	if boundary == "" {
		return attachments, nil
	}

	attDir := filepath.Join(outputDir, attachmentsDir)

	mr := multipart.NewReader(msg.Body, boundary)
	return extractMultipartAttachments(mr, attDir)
}

func extractMultipartAttachments(mr *multipart.Reader, attDir string) ([]Attachment, error) {
	var attachments []Attachment

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		// Check for nested multipart
		partContentType := part.Header.Get("Content-Type")
		partMediaType, partParams, _ := mime.ParseMediaType(partContentType)
		if strings.HasPrefix(partMediaType, "multipart/") {
			if boundary := partParams["boundary"]; boundary != "" {
				nestedMr := multipart.NewReader(part, boundary)
				if nested, err := extractMultipartAttachments(nestedMr, attDir); err == nil {
					attachments = append(attachments, nested...)
				}
			}
			continue
		}

		disposition := part.Header.Get("Content-Disposition")
		if !strings.Contains(disposition, "attachment") {
			continue
		}

		filename := part.FileName()
		if filename == "" {
			// Try to extract filename from Content-Disposition
			_, params, err := mime.ParseMediaType(disposition)
			if err == nil {
				filename = params["filename"]
			}
		}

		if filename == "" {
			continue
		}

		filename = decodeHeader(filename)
		filename = sanitizeFilename(filename)

		// Create attachments directory
		if err := os.MkdirAll(attDir, 0755); err != nil {
			return attachments, err
		}

		// Handle duplicate filenames
		fullPath := filepath.Join(attDir, filename)
		fullPath = makeUniqueFilepath(fullPath)
		filename = fullPath[len(attDir)+1:] // Get just the filename

		// Save attachment
		content, err := io.ReadAll(part)
		if err != nil {
			fmt.Fprintf(os.Stderr, "   Warning: Failed to read attachment %s: %v\n", filename, err)
			continue
		}

		// Decode base64 if needed (Content-Transfer-Encoding)
		encoding := part.Header.Get("Content-Transfer-Encoding")
		if encoding == "base64" {
			decoded := make([]byte, base64.StdEncoding.DecodedLen(len(content)))
			n, err := base64.StdEncoding.Decode(decoded, content)
			if err == nil {
				content = decoded[:n]
			}
		} else if encoding == "quoted-printable" {
			qpr := quotedprintable.NewReader(bytes.NewReader(content))
			if decoded, err := io.ReadAll(qpr); err == nil {
				content = decoded
			}
		}

		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "   Warning: Failed to save attachment %s: %v\n", filename, err)
			continue
		}

		fileInfo, _ := os.Stat(fullPath)
		size := fileInfo.Size()

		attachments = append(attachments, Attachment{
			Filename: filename,
			Path:     filepath.Join(attachmentsDir, filename),
			Size:     size,
		})

		fmt.Fprintf(os.Stderr, "   Extracted attachment: %s (%d bytes)\n", filename, size)
	}

	return attachments, nil
}

func createEmailMarkdown(metadata EmailMetadata, body string, attachments []Attachment) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("# Email: %s\n\n", metadata.Subject))
	md.WriteString("## Metadata\n\n")
	md.WriteString(fmt.Sprintf("- **From:** %s\n", metadata.From))

	if len(metadata.To) > 0 {
		md.WriteString(fmt.Sprintf("- **To:** %s\n", strings.Join(metadata.To, ", ")))
	}

	if len(metadata.Cc) > 0 {
		md.WriteString(fmt.Sprintf("- **Cc:** %s\n", strings.Join(metadata.Cc, ", ")))
	}

	md.WriteString(fmt.Sprintf("- **Date:** %s\n", metadata.Date))
	md.WriteString(fmt.Sprintf("- **Subject:** %s\n\n", metadata.Subject))

	// Add attachments section
	if len(attachments) > 0 {
		md.WriteString("## Attachments\n\n")
		for _, att := range attachments {
			size := formatFileSize(att.Size)
			md.WriteString(fmt.Sprintf("- **%s** (%s) - `%s`\n", att.Filename, size, att.Path))
		}
		md.WriteString("\n")
	}

	md.WriteString("---\n\n")
	md.WriteString("## Message\n\n")
	md.WriteString(body)
	md.WriteString("\n\n---\n")

	// Add thread information if available
	if metadata.InReplyTo != "" || metadata.References != "" {
		md.WriteString("\n## Thread Information\n\n")
		if metadata.MessageID != "" {
			md.WriteString(fmt.Sprintf("- **Message ID:** `%s`\n", metadata.MessageID))
		}
		if metadata.InReplyTo != "" {
			md.WriteString(fmt.Sprintf("- **In Reply To:** `%s`\n", metadata.InReplyTo))
		}
		if metadata.References != "" {
			md.WriteString(fmt.Sprintf("- **References:** `%s`\n", metadata.References))
		}
	}

	return md.String()
}

func printExtractionSummary(result *ExtractionResult) {
	fmt.Fprintf(os.Stderr, "\n📧 Email: %s\n", result.Metadata.Subject)
	fmt.Fprintf(os.Stderr, "📁 Output directory: %s\n", result.OutputDir)
	fmt.Fprintf(os.Stderr, "📝 Markdown file: %s\n", result.MarkdownFile)
	fmt.Fprintf(os.Stderr, "📎 Attachments extracted: %d\n", len(result.Attachments))
	if len(result.Attachments) > 0 {
		fmt.Fprintf(os.Stderr, "   Attachment files:\n")
		for _, att := range result.Attachments {
			size := formatFileSize(att.Size)
			fmt.Fprintf(os.Stderr, "   - %s (%s)\n", att.Filename, size)
		}
	}
	fmt.Fprintf(os.Stderr, "\n%s\n", strings.Repeat("=", 80))
	fmt.Fprintf(os.Stderr, "EXTRACTED CONTENT:\n")
	fmt.Fprintf(os.Stderr, "%s\n", strings.Repeat("=", 80))
	fmt.Println(result.Markdown)
	fmt.Fprintf(os.Stderr, "%s\n", strings.Repeat("=", 80))
}

func cleanupExtraction(outputDir string) {
	if err := os.RemoveAll(outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "\n⚠️  Warning: Failed to clean up: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "\n🧹 Cleaned up: %s\n", outputDir)
	}
}

// Helper functions

func decodeHeader(s string) string {
	dec := new(mime.WordDecoder)
	decoded, err := dec.DecodeHeader(s)
	if err != nil {
		return s
	}
	return decoded
}

func formatEmailAddress(addr string) string {
	if addr == "" {
		return "Unknown"
	}

	addresses, err := mail.ParseAddressList(addr)
	if err != nil || len(addresses) == 0 {
		return addr
	}

	if addresses[0].Name != "" {
		return fmt.Sprintf("%s <%s>", addresses[0].Name, addresses[0].Address)
	}
	return addresses[0].Address
}

func parseAddressList(addrs string) []string {
	if addrs == "" {
		return nil
	}

	addresses, err := mail.ParseAddressList(addrs)
	if err != nil {
		// Fallback to simple split
		parts := strings.Split(addrs, ",")
		var result []string
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}

	var result []string
	for _, addr := range addresses {
		if addr.Name != "" {
			result = append(result, fmt.Sprintf("%s <%s>", addr.Name, addr.Address))
		} else {
			result = append(result, addr.Address)
		}
	}
	return result
}

func htmlToText(htmlContent string) string {
	// Remove script and style elements
	reScript := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	htmlContent = reScript.ReplaceAllString(htmlContent, "")
	reStyle := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	htmlContent = reStyle.ReplaceAllString(htmlContent, "")

	// Convert common HTML tags
	reBr := regexp.MustCompile(`(?i)<br\s*/?>`)
	htmlContent = reBr.ReplaceAllString(htmlContent, "\n")
	reP := regexp.MustCompile(`(?i)</p>`)
	htmlContent = reP.ReplaceAllString(htmlContent, "\n\n")
	reDiv := regexp.MustCompile(`(?i)</div>`)
	htmlContent = reDiv.ReplaceAllString(htmlContent, "\n")
	reTr := regexp.MustCompile(`(?i)</tr>`)
	htmlContent = reTr.ReplaceAllString(htmlContent, "\n")
	reLi := regexp.MustCompile(`(?i)<li>`)
	htmlContent = reLi.ReplaceAllString(htmlContent, "\n- ")
	reH := regexp.MustCompile(`(?i)</h[1-6]>`)
	htmlContent = reH.ReplaceAllString(htmlContent, "\n\n")

	// Remove all remaining HTML tags
	reTags := regexp.MustCompile(`<[^>]+>`)
	htmlContent = reTags.ReplaceAllString(htmlContent, "")

	// Unescape HTML entities
	htmlContent = html.UnescapeString(htmlContent)

	// Clean up whitespace
	lines := strings.Split(htmlContent, "\n")
	var cleaned []string
	prevBlank := false
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		isBlank := strings.TrimSpace(trimmed) == ""
		if !(isBlank && prevBlank) {
			cleaned = append(cleaned, trimmed)
		}
		prevBlank = isBlank
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

func sanitizeFilename(name string) string {
	// Replace invalid characters
	re := regexp.MustCompile(`[^\w\s\-.]`)
	name = re.ReplaceAllString(name, "_")
	// Replace multiple spaces with single underscore
	reSpaces := regexp.MustCompile(`\s+`)
	name = reSpaces.ReplaceAllString(name, "_")
	return name
}

func makeUniqueFilepath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	dir := filepath.Dir(path)
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	counter := 1
	for {
		newFilename := fmt.Sprintf("%s_%d%s", base, counter, ext)
		newPath := filepath.Join(dir, newFilename)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absPath
}

func decodeContent(content []byte, encoding string) ([]byte, error) {
	reader, err := charset.NewReaderLabel(encoding, bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	return io.ReadAll(reader)
}
