#!/usr/bin/env perl
use v5.14;
use POSIX;

use Cat;
use Cli;
use PushID;

main();

sub main {
  my %switch = Cli::read_switches();
  my ($dt, $cat, $amt, $body) = read_params();

  add_expense($dt, $cat, $amt, $body);

  # -v (verbose): Write summary of newly added expense.
  if (exists $switch{"v"}) {
    my $catDesc = Cat::desc($cat);
    my $samt = sprintf("%.2f", $amt);

    say "$dt $catDesc";
    say "$samt";
    say "$body";
  }
}

sub read_params {
  my $dt;
  my $cat;
  my $amt;
  my $body;

  for my $arg (@ARGV) {
    if ($arg =~ /^-\w+$/) {
      # Skip over switches
      next;
    } elsif ($arg =~ /^--\w+$/) {
      next;
    } elsif ($arg =~ /^--\w+=\w+$/) {
      next;
    }elsif (!$cat && Cat::exists($arg)) {
      # category id
      $cat = $arg;
    } elsif (!$dt && $arg =~ /^\d\d\d\d-\d\d-\d\d$/) {
      # date: YYYY-MM-DD
      $dt = $arg;
    } elsif (!$amt && $arg =~ /^\d+\.\d+$/) {
      # amt: nnn.nn
      $amt = $arg + 0;
    } elsif (!$body) {
      # body: any text
      $body = $arg;
    }
  }

  # Use current date if none specified.
  if (!$dt) {
    $dt = strftime("%Y-%m-%d", localtime(time));
  }

  if (!$cat) {
    my $csvCatIds = join(", ", Cat::ids());
    say STDERR "Please specify a category\n(Ex. $csvCatIds)";
    exit(1);
  }

  if (!$amt) {
    say STDERR "Please specify an amount (Ex. 123.45)";
    exit(1);
  }

  if (!$body) {
    $body = "";
  }

  return ($dt, $cat, $amt, $body);
}

sub add_expense {
  my ($dt, $cat, $amt, $body) = @_;

  my $datafile = "$ENV{'HOME'}/.expbuddata";
  open (DATA, ">>", $datafile) || die "Error opening $datafile for output.";
  my $id = PushID::gen_id();
  my $samt = sprintf("%.2f", $amt);
  my $line = "$id|$dt|$cat|$samt|$body";
  say DATA $line;
  close DATA;
}

