# Notes Buddy GUI

Tcl scripts for Notes Buddy GUI. Requires Tcl/Tk environment.

Notes are stored in _~/.nbdata_ file.

Each 'note' is stored in the following format:

```
Line 0:
<id>|<date>|<tags>

Line 1..n
<body> section

Ex.
1|2018-08-01|tag1, multi-word tag, tag2
This is the note body. It can be comprised of multiple paragraphs separated by a blank line.

Paragraph 2 here, more words.

Another paragraph. This note will be terminated until the fields csv section of the next note begins.
```

## Usage:

```
./nb-gui.tcl
```

## Update: I've revised the 'notes' file so it can accomodate any kind of field. Instead of a fixed csv line for the records, you can use dot commands to define fields. Sample below:

```
%%<id>
.title Title of note
.tags tag1,tag2,multi word tag,tag3
Note body appears here. It can be comprised of multiple paragraphs.

Paragraph 2 here, more words to come.

.field1 You can also define fields anywhere in the note by starting the line with a dot '.' followed by field name, then value of the field. The .field definition will be excluded from the body text.

Paragraph 3 resumes here.
```

Code for the new 'notes' file format is in nb.tcl . I will revise the GUI code (nb-gui.tcl) to use the new notes format soon.

