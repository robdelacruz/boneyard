#!/usr/bin/env perl
use v5.14;

use Cat;
use Cli;
use PushID;

main();

sub main {
  my %switch = Cli::read_switches();

  my ($ids, $dt, $cat, $amt, $body) = read_params();

  if (@$ids > 0) {
    edit_expense(\%switch, $ids, $dt, $cat, $amt, $body);
  }
}

sub read_params() {
  my @ids;
  my $dt;
  my $cat;
  my $amt;
  my $body;

  for my $arg (@ARGV) {
    if (PushID::is_id($arg)) {
      push @ids, $arg;
    } elsif (!$cat and Cat::exists($arg)) {
      $cat = $arg;
    } elsif (!$dt and $arg =~ /^\d\d\d\d-\d\d-\d\d$/) {
      $dt = $arg;
    } elsif (!$amt and $arg =~ /^\d+\.\d+$/) {
      $amt = $arg + 0;
    } elsif (!$body) {
      $body = $arg;
    }
  }

  # Also read ids from stdin if any were piped in.
  # Query results can be piped to edit to edit all query matches.
  # Ex. expb-q <search str> --showid | expb-edit <fields to change>
  if (not -t STDIN) {   # only when stdin not coming from terminal
    while (<STDIN>) {
      my ($id) = split(/[\|\s]/, $_);

      if (PushID::is_id($id)) {
        push @ids, $id;
      }
    }
  }

  return (\@ids, $dt, $cat, $amt, $body);
}

sub edit_expense {
  my($switch, $ids, $newdt, $newcat, $newamt, $newbody) = @_;

  if (@$ids == 0) {
    return;
  }

  my $datafile = "$ENV{'HOME'}/.expbuddata";
  open (DATA, "<", $datafile) || die "Error opening $datafile for read.";

  my $tmpfile = "$ENV{'HOME'}/.expbuddata.tmp";
  open (OUT, ">", $tmpfile) || die "Error opening $tmpfile for write.";

  while (my $line = <DATA>) {
    chomp $line;

    for my $id (@$ids) {
      if ($line =~ /^$id\|/) {
        $line = replace_line_fields($line, $id, $newdt, $newcat, $newamt, $newbody);

        # -v (verbose): Write summary of edited expense.
        if (exists $switch->{"v"}) {
          say $line;
        }

        last;
      }
    }
    say OUT $line;
  }

  close DATA;
  close OUT;

  rename $datafile, "$ENV{'HOME'}/.expbuddata.bak";
  rename $tmpfile, $datafile;
  unlink $tmpfile;
}

sub replace_line_fields {
  my($line, $id, $newdt, $newcat, $newamt, $newbody) = @_;

  my (undef, $dt, $cat, $amt, $body) = split(/\|/, $line);
  $dt = $newdt if ($newdt);
  $cat = $newcat if ($newcat);
  $amt = sprintf("%.2f", $newamt) if ($newamt);
  $body = $newbody if ($newbody);

  $line = "$id|$dt|$cat|$amt|$body";

  return $line;

}


