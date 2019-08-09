#!/usr/bin/env perl
use v5.14;
use POSIX;

use Cat;
use Cli;
use PushID;

# Parameters
# ----------
# YYYY
# YYYY-MM
# YYYY-MM-DD
# YYYY-MM-DD...
# ...YYYY-MM-DD
# YYYY-MM-DD...YYYY-MM-DD
# cat:<cat1,cat2,...>
# <any text>  (searches all partial matches in body field)
#
# Switches
# --------
# -t Show totals
# --ytd Year to date (same as YYYY of current year)
# --mtd Month to date (same as YYYY-MM of current year and month)
# --dtd Day to date (same as YYYY-MM-DD of current day)
# --recentdays=<n> Show expenses from past n days.

main();

sub main {
  my %switch = Cli::read_switches();
  my ($qid, $dtmin, $dtmax, $qcatIds, $qbody) = read_params(\%switch);

  unless (defined $qid or defined $dtmin or defined $dtmax or hashLen($qcatIds) or defined $qbody) {
    my $usage = <<EOM;
Usage
-----
Combine one or more of the following:
  (all expenses in given year)
  YYYY (Ex. '2018')

  (all expenses in a given year and month)
  YYYY-MM (Ex. '2018-07')

  (all expenses in date)
  YYYY-MM-DD (Ex. '2018-07-13')

  (all expense on or after a date)
  YYYY-MM-DD... (Ex. '2018-07-01...')

  (all expenses on or before a date)
  ...YYYY-MM-DD (Ex. '...2018-07-31')

  (all expenses in date range)
  YYYY-MM-DD...YYYY-MM-DD (Ex. '2018-07-01...2018-07-31')

  (all expenses in specific categories)
  cat:<category1,category2,...> (Ex. 'cat:commute,food,household')

  (search body text)
  <any text>
EOM
    say STDERR $usage;
    exit(1);
  }

  my @results = query_expenses($qid, $dtmin, $dtmax, $qcatIds, $qbody);

  # Sort results by date+id descending.
  @results = sort { cmp_expline_date_id($a, $b); } @results;
  write_results(\@results, \%switch);
}

sub hashLen {
  my ($hash) = @_;
  return scalar keys %$hash;
}

sub read_params {
  my ($switch) = @_;

  my $id;
  my $dtmin;
  my $dtmax;
  my %qcatIds;
  my $qbody;

  for my $arg (@ARGV) {
    if (!$id and PushID::is_id($arg)) {
      $id = $arg;
    } elsif ($arg =~ /^\d\d\d\d$/) {
      # YYYY
      $dtmin = "$arg-01-01";
      $dtmax = "$arg-12-31";
    } elsif ($arg =~ /^\d\d\d\d-\d\d$/) {
      # YYYY-MM
      $dtmin = "$arg-01";
      $dtmax = "$arg-31";
    } elsif ($arg =~ /^\d\d\d\d-\d\d-\d\d$/) {
      # YYYY-MM-DD
      $dtmin = $arg;
      $dtmax = $arg;
    } elsif ($arg =~ /^\d\d\d\d-\d\d-\d\d\.\.\.$/) {
      # YYYY-MM-DD...
      ($dtmin) = $arg =~ /^(\d\d\d\d-\d\d-\d\d)\.\.\.$/
    } elsif ($arg =~ /^\.\.\.\d\d\d\d-\d\d-\d\d$/) {
      # ...YYYY-MM-DD
      ($dtmax) = $arg =~ /^\.\.\.(\d\d\d\d-\d\d-\d\d)$/
    } elsif ($arg =~ /^\d\d\d\d-\d\d-\d\d\.\.\.\d\d\d\d-\d\d-\d\d$/) {
      # YYYY-MM-DD...YYYY-MM-DD
      ($dtmin, $dtmax) =
        $arg =~ /^(\d\d\d\d-\d\d-\d\d)\.\.\.(\d\d\d\d-\d\d-\d\d)$/
    } elsif ($arg =~ /^cat:[\w,]+$/) {
      # cat:<catId1>,<catId2>...
      my ($csv) = $arg =~ /^cat:([\w,]+)$/;
      my @ids = split(/,/, $csv);
      %qcatIds = map {$_ => 1} @ids;
    } else {
      # search any leftover text in body field
      $qbody = $arg;
    }
  }

  if (exists $switch->{"ytd"}) {
    my $today = time;
    $dtmin = strftime("%Y", localtime($today)) . "-01-01";
    $dtmax = strftime("%Y", localtime($today)) . "-12-31";
  }
  if (exists $switch->{"mtd"}) {
    my $today = time;
    $dtmin = strftime("%Y-%m", localtime($today)) . "-01";
    $dtmax = strftime("%Y-%m", localtime($today)) . "-31";
  }
  if (exists $switch->{"dtd"}) {
    my $today = time;
    $dtmin = strftime("%Y-%m-%d", localtime($today));
    $dtmax = strftime("%Y-%m-%d", localtime($today));
  }
  if (exists $switch->{"recentdays"}) {
    my $ndays = $switch->{"recentdays"} || 1;
    my $today = time;
    my $prevday = $today - 24*60*60*($ndays-1);
    $dtmin = strftime("%Y-%m-%d", localtime($prevday));
    $dtmax = strftime("%Y-%m-%d", localtime($today));
  }

  return ($id, $dtmin, $dtmax, \%qcatIds, $qbody);
}

sub query_expenses {
  my ($qid, $dtmin, $dtmax, $qcatIds, $qbody) = @_;

  my $datafile = "$ENV{'HOME'}/.expbuddata";
  open (DATA, "<", $datafile) || die "Error opening $datafile for reading.";

  my @results;
  while (my $line = <DATA>) {
    chomp $line;
    my ($id, $dt, $cat, $amt, $body) = split(/\|/, $line);

    if (defined $qid and $qid ne $id) {
      next;
    }
    if (defined $dtmin and $dt lt $dtmin) {
      next;
    }
    if (defined $dtmax and $dt gt $dtmax) {
      next;
    }
    if (hashLen($qcatIds) and not exists $qcatIds->{$cat}) {
      next;
    }
    if (defined $qbody and not $body =~ /\Q$qbody\E/i) {
      next;
    }

    push @results, $line;
  }

  close DATA;
  return @results;
}

# order by date and id (descending order of date and entry time)
sub cmp_expline_date_id {
  my ($l1, $l2) = @_;

  my ($id1, $dt1) = split(/\|/, $l1);
  my ($id2, $dt2) = split(/\|/, $l2);
  my $s1 = "$dt1 $id1";
  my $s2 = "$dt2 $id2";

  return $s2 cmp $s1;

}

sub write_results {
  my ($results, $switch) = @_;

  my @results = @$results;

  # Get each field max len for column alignment purposes.
  my $idlen = 0;
  my $dtlen = 0;
  my $catlen = 0;
  my $amtlen = 0;
  my $bodylen = 0;
  for (@results) {
    chomp;
    my ($id, $dt, $cat, $amt, $body) = split(/\|/, $_);

    if (length $id > $idlen) {
      $idlen = length $id;
    }
    if (length $dt > $dtlen) {
      $dtlen = length $dt;
    }
    my $catDesc = Cat::desc($cat);
    if (length $catDesc > $catlen) {
      $catlen = length $catDesc;
    }
    my $samt = sprintf("%.2f", $amt);
    if (length $samt > $amtlen) {
      $amtlen = length $samt;
    }
    if (length $body > $bodylen) {
      $bodylen = length $body;
    }
  }

  # Write row results.
  my $amtTotal = 0.0;
  for (@results) {
    chomp;
    my ($id, $dt, $cat, $amt, $body) = split(/\|/, $_);

    my $samt = sprintf("%.2f", $amt);
    my $catDesc = Cat::desc($cat);

    if (exists $switch->{"showid"}) {
      say sprintf("%-${idlen}s  %-${dtlen}s  %-${catlen}s  %${amtlen}s  %-${bodylen}s", $id, $dt, $catDesc, $samt, $body);
    } else {
      say sprintf("%-${dtlen}s  %-${catlen}s  %${amtlen}s  %-${bodylen}s", $dt, $catDesc, $samt, $body);
    }

    $amtTotal += $amt;
  }

# -t switch: Show Totals
  if (exists $switch->{"t"} and @results > 0) {
    say sprintf("%-${idlen}s--%-${dtlen}s--%-${catlen}s--%${amtlen}s--%-${bodylen}s",
      "-" x $idlen, "-" x $dtlen, "-" x $catlen, "-" x $amtlen, "-" x $bodylen);

    my $samtTotal = sprintf("%.2f", $amtTotal);

    if (exists $switch->{"showid"}) {
      say sprintf("%-${idlen}s  %-${dtlen}s  %-${catlen}s  %${amtlen}s  %-${bodylen}s", "", "", "Total:", $samtTotal, "");
    } else {
      say sprintf("%-${dtlen}s  %-${catlen}s  %${amtlen}s  %-${bodylen}s", "", "Total:", $samtTotal, "");
    }
  }
}

