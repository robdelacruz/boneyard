package Cli;

use v5.14;

sub read_switches {
  my %switch;

  my $i = 0;
  while ($i <= $#ARGV) {
    my $arg = $ARGV[$i];

    # $$ Note: If arg length >= 20, don't consider it a single char switch.
    #          This is a hack to handle guid params that begin with dash '-'.
    if (length $arg < 20 and $arg =~ /^-(\w+)$/) {
      # -abc  single char options a, b, c
      for my $ch (split //, $1) {
        $switch{$ch} = 1;
      }

      # Remove switch from args
      splice @ARGV, $i, 1;
      next;
    } elsif ($arg =~ /^--(\w+)(?:=(\w+))?$/) {
      # --opt=val  val is optional
      $switch{$1} = $2;

      # Remove switch from args
      splice @ARGV, $i, 1;
      next;
    }
    $i++;
  }

  return %switch;
}

1;

