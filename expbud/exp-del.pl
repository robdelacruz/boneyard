#!/usr/bin/env perl
use v5.14;

use Cat;
use Cli;
use PushID;

main();

sub main {
  my %switch = Cli::read_switches();

  my ($ids) = read_params();

  if (@$ids > 0) {
    del_expense(\%switch, $ids);
  }
}

sub read_params() {
  my @ids;

  for my $arg (@ARGV) {
    if (PushID::is_id($arg)) {
      push @ids, $arg;
    }
  }

  # Also read ids from stdin if any were piped in.
  # Query results can be piped to edit to edit all query matches.
  # Ex. expb-q <search str> --showid | expb-del
  if (not -t STDIN) {   # only when stdin not coming from terminal
    while (<STDIN>) {
      my ($id) = split(/[\|\s]/, $_);

      if (PushID::is_id($id)) {
        push @ids, $id;
      }
    }
  }

  return (\@ids);
}

sub del_expense {
  my($switch, $ids) = @_;

  if (@$ids == 0) {
    return;
  }

  my $datafile = "$ENV{'HOME'}/.expbuddata";
  open (DATA, "<", $datafile) || die "Error opening $datafile for read.";

  my $tmpfile = "$ENV{'HOME'}/.expbuddata.tmp";
  open (OUT, ">", $tmpfile) || die "Error opening $tmpfile for write.";

LINES:
  while (my $line = <DATA>) {
    chomp $line;

    for my $id (@$ids) {
      if ($line =~ /^$id\|/) {
        # -v (verbose): Write summary of deleted expense.
        if (exists $switch->{"v"}) {
          say "Deleted: $line";
        }

        # Skip writing this line to be deleted.
        next LINES;
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


