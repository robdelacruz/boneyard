use v5.14;

# $ buildsite.pl *.md
# Converts all *.md markdown files into html
# Use page.tpl as template for all pages.

# Read template
my $tpl = `cat page.tpl`;

# Each input file (ex. *.md)
for my $file (@ARGV) {
  unless (-e $file) {
    next;
  }
  open(my $fh, "<", $file) || next;

  # Page title is first nonempty line of input file.
  my $title;
  while ($title = <$fh>) {
    chomp($title);

    # Remove left "# " and trim spaces.
    $title =~ s/^[\s#]+|\s+$//g;
    last if $title;
  }
  close($fh);

  # Convert markdown text to html
  my $body = `cat $file | Markdown.pl`;

  # Fill in template.
  my $html = $tpl;
  $html =~ s/{title}/$title/;
  $html =~ s/{body}/$body/;

  # Output filename
  `mkdir -p dist`;
  my $outfile = "dist/$file";
  $outfile =~ s/\..*$/\.html/;

  # Write html to output file.
  open(my $fh, ">", $outfile) || die "Error opening $outfile for output.";
  print $fh $html;
  close $fh;
}

# Copy any images to dist/
`cp *.gif dist/ 2>/dev/null`;
`cp *.png dist/ 2>/dev/null`;
`cp *.jpg dist/ 2>/dev/null`;

