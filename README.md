# email-extractor

A Go CLI tool that extracts content and attachments from `.eml` email files and converts them to structured markdown format.

## Features

- 📧 **Email Parsing** - Extracts metadata (From, To, Cc, Subject, Date, etc.)
- 📝 **Markdown Output** - Converts email content to readable markdown
- 📎 **Attachment Extraction** - Saves all attachments to organized directories
- 🌐 **Encoding Support** - Handles various character encodings (UTF-8, ISO-8859-1, etc.)
- 🔄 **HTML Conversion** - Converts HTML emails to clean text
- 🗂️ **MIME Handling** - Processes multipart and nested multipart messages
- 🧹 **Cleanup Mode** - Optional automatic cleanup after extraction

## Installation

### From Source

```bash
# Clone or navigate to the project directory
cd email-extractor

# Build and install
make install

# Or install to custom location
TARGET=~/bin make install
```

### Manual Build

```bash
cd src
go build -o email-extractor .
```

## Usage

### Basic Extraction

```bash
email-extractor message.eml
```

This creates a folder named `{subject}_email/` containing:
- `email.md` - Markdown file with email content and metadata
- `attachments/` - Directory with all extracted attachments

### Custom Output Directory

```bash
email-extractor ~/Downloads/email.eml ~/Documents/extracted
```

### Auto-Cleanup Mode

Extract and display content, then automatically delete the extraction folder:

```bash
email-extractor --cleanup message.eml
```

### Command-Line Options

```
Usage: email-extractor [options] <eml_file> [output_directory]

Arguments:
  eml_file           Path to the .eml file to extract
  output_directory   Optional: Base directory for extraction
                     Default: Same directory as .eml file

Options:
  --cleanup          Clean up extraction directory after reading
```

## Output Format

### Directory Structure

```
{subject}_email/
├── email.md           # Main markdown file
└── attachments/       # Extracted attachments
    ├── document.pdf
    ├── image.png
    └── ...
```

### Markdown Content

The generated `email.md` includes:

1. **Email Header** - Subject as title
2. **Metadata Section** - From, To, Cc, Date, Subject
3. **Attachments List** - Filenames, sizes, and paths
4. **Message Body** - Email content (converted from HTML if needed)
5. **Thread Information** - Message-ID, In-Reply-To, References (if available)

### Example Output

```markdown
# Email: Meeting Notes - Q4 Planning

## Metadata

- **From:** John Doe <john@example.com>
- **To:** team@example.com
- **Date:** 2025-12-14 10:30:00
- **Subject:** Meeting Notes - Q4 Planning

## Attachments

- **presentation.pdf** (2.3 MB) - `attachments/presentation.pdf`
- **budget.xlsx** (145.2 KB) - `attachments/budget.xlsx`

---

## Message

[Email content here...]
```

## Development

### Build Commands

```bash
# Build binary
make build

# Install to /usr/local/bin
make install

# Uninstall
make uninstall

# Clean build artifacts
make clean

# Update dependencies
make go.sum
```

### Project Structure

```
email-extractor/
├── src/
│   ├── main.go           # Main application code
│   ├── go.mod            # Go module definition
│   └── go.sum            # Dependency checksums
├── Makefile              # Build automation
├── CLAUDE.md             # AI-oriented documentation
└── README.md             # This file
```

## Dependencies

- **Go 1.21+**
- **golang.org/x/net** - HTML character set detection and conversion

## Technical Details

### Supported Email Formats

- Simple text emails
- HTML emails (converted to text)
- Multipart/alternative (HTML + text)
- Multipart/mixed (with attachments)
- Nested multipart messages
- Quoted-printable and base64 encodings

### Character Encoding

Automatically detects and converts various character encodings:
- UTF-8
- ISO-8859-1 (Latin-1)
- Windows-1252
- And many others via `golang.org/x/net/html/charset`

### Attachment Handling

- Decodes base64 and quoted-printable encodings
- Handles duplicate filenames (appends `_1`, `_2`, etc.)
- Sanitizes filenames (removes invalid characters)
- Preserves original file extensions

## Examples

### Extract Email from Downloads

```bash
email-extractor ~/Downloads/important_email.eml
```

Output: `~/Downloads/Important_Message_email/email.md`

### Extract and Review, Then Cleanup

```bash
# Extract and display
email-extractor message.eml

# Review the content (printed to stdout)
# Manually delete when done
rm -rf {subject}_email/

# Or use auto-cleanup mode
email-extractor --cleanup message.eml
```

### Process Multiple Emails

```bash
for eml in ~/Downloads/*.eml; do
  email-extractor "$eml"
done
```

## Troubleshooting

### "Email file not found"
- Check that the path to the `.eml` file is correct
- Use absolute paths or `~/` for home directory

### "Failed to parse email"
- Ensure the file is a valid `.eml` format
- Try opening the file in an email client to verify it's not corrupted

### Attachments Not Extracted
- Check that the email actually contains attachments
- Some inline images may be embedded as base64 in HTML rather than attachments

### Character Encoding Issues
- The tool auto-detects encodings, but some rare encodings may not be supported
- Check the original email in a desktop email client

## License

This tool is provided as-is for email content extraction purposes.

## Contributing

This is a personal utility project. For suggestions or issues, please contact the maintainer.

## Author

Sebastien MORAND - sebastien.morand@loreal.com
