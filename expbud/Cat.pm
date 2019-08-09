package Cat;

use v5.14;

my %_catDesc = (
  none => "None",
  commute => "Commute",
  dineout => "Dine Out",
  grocery => "Grocery",
  food => "Food",
  utilities => "Utilities",
  household => "Household",
);

sub desc {
  my ($cat) = @_;

  return $_catDesc{$cat} || $cat;
}

sub exists {
  my ($cat) = @_;

  return exists $_catDesc{$cat};
}

sub ids {
  return sort(keys %_catDesc);
}

1;
