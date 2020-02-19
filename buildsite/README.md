# buildsite

Rob's personal website generator and content source files.

Contains:

- Script for generating markup files from markdown.
- Source markdown files *.md representing web pages.

Requires Perl 5 and [Markdown.pl](http://daringfireball.net/projects/downloads/Markdown_1.0.1.zip) in bin path.

## Usage

```
$ perl buildsite.pl *.md
```

This will:

- Generates html into dist/ 
- Copies all images into dist/

## License

MIT

