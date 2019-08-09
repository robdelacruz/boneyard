# expbud - Expense Buddy command line

Perl scripts for Expense Buddy command line version.

Expenses are stored as "|" csv fields in _~/.expbuddata_ file.

## Usage:


Add a new expense:

```
./expb-add.pl <category> <amt> [body] [date]

<category>: Category alias
<amt>: Expense Amount (Ex. 123.45)
[body]: Expense Description (Ex. Lunch at McDonald's)
[date]: ISO 8601 Date format YYYY-MM-DD (Ex. 2018-06-22)

If _[date]_ not specified, defaults to today's date.

```
Query expenses:
```
./expb-q.pl [switches] [date range] [cat:categories...] [search str] 

Query for the following:
YYYY
YYYY-MM
YYYY-MM-DD
YYYY-MM-DD...
...YYYY-MM-DD
YYYY-MM-DD...YYYY-MM-DD
cat:<cat1,cat2,...>
<any text>  (searches all partial matches in body field)

Switches:
-t Show totals
--ytd Year to date (same as YYYY of current year)
--mtd Month to date (same as YYYY-MM of current year and month)
--dtd Day to date (same as YYYY-MM-DD of current day)
--recentdays=<n> Show expenses from past n days.

```


