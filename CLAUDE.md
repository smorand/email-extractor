# email-extractor - AI Documentation

## Overview

**email-extractor** is a Go CLI tool that extracts content and attachments from `.eml` email files and converts them to structured markdown format. It handles multipart MIME messages, decodes various character encodings, and preserves email metadata.

## Project Structure

```
email-extractor/
├── src/
│   ├── main.go           # Entry point and core extraction logic
│   ├── go.mod            # Go module definition
│   └── go.sum            # Dependency checksums
├── Makefile              # Build automation
├── CLAUDE.md             # AI-oriented documentation (this file)
└── README.md             # Human-oriented documentation
```

## Core Components

### Data Structures

Located in `src/main.go:27-54`:

- **EmailMetadata** - Parsed email headers (From, To, Cc, Subject, Date, Message-ID, etc.)
- **Attachment** - Attachment metadata (filename, path, size)
- **ExtractionResult** - Complete extraction output including markdown, metadata, and attachments

### Main Functions

- **main()** (`src/main.go:56`) - CLI entry point, parses flags and orchestrates extraction
- **extractEmailContent()** (`src/main.go:109`) - Main extraction workflow
- **parseEmailMetadata()** (`src/main.go:201`) - Extracts email headers
- **getEmailBody()** (`src/main.go:228`) - Extracts email message body (text/HTML)
- **extractAttachments()** (`src/main.go:344`) - Extracts and saves attachments
- **createEmailMarkdown()** (`src/main.go:462`) - Generates markdown output

### Key Features

1. **Multipart MIME Handling** - Recursively processes nested multipart messages
2. **Character Encoding** - Decodes various charsets (UTF-8, ISO-8859-1, etc.)
3. **HTML to Text** - Converts HTML email bodies to readable text (`src/main.go:595`)
4. **Attachment Extraction** - Handles base64 and quoted-printable encodings
5. **Cleanup Mode** - Optional automatic cleanup of extraction directory

## Usage Patterns

### Basic Extraction
```bash
email-extractor message.eml
```

### Custom Output Directory
```bash
email-extractor ~/Downloads/email.eml ~/Documents/extracted
```

### Auto-cleanup
```bash
email-extractor --cleanup message.eml
```

## Build & Install

```bash
# Build binary
make build

# Install to /usr/local/bin
make install

# Install to custom location
TARGET=~/bin make install

# Uninstall
make uninstall
```

## Dependencies

- **golang.org/x/net/html/charset** - Character encoding detection and conversion
- Standard library: `mime`, `mime/multipart`, `net/mail`, `encoding/base64`, etc.

## Code Organization Compliance

### Standards Followed
- Single responsibility functions
- Error wrapping with `%w`
- Constants for magic strings
- Clear naming conventions
- Alphabetically ordered fields in structs

### Known Deviations
- All code in single file (acceptable for small CLI tools)
- Helper functions grouped at end rather than by call order
- No separate packages (not needed for this scope)

## Extending the Tool

### Adding New Output Formats

Modify `createEmailMarkdown()` or add new functions like `createEmailJSON()`, `createEmailHTML()`.

### Custom Attachment Handling

Extend `extractMultipartAttachments()` to filter or process specific file types.

### Additional Metadata

Add fields to `EmailMetadata` struct and update `parseEmailMetadata()`.

## Testing Considerations

- Test with various email clients (Gmail, Outlook, Thunderbird)
- Test multipart/alternative, multipart/mixed, multipart/related
- Test various character encodings
- Test large attachments and edge cases (duplicate filenames, special characters)
- Test nested multipart messages (common in forwarded emails)

## Security Notes

- File paths are sanitized to prevent directory traversal
- Filenames are sanitized to remove invalid characters
- No execution of email content (safe static extraction only)
