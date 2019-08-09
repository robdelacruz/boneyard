#!/usr/bin/env wish

#
# GUI section
#

proc activateSection {section} {
  if {$section eq "search"} {
    pack .1 -side top -fill both -expand true
    pack forget .2
  } else {
    pack .2 -side top -fill both -expand true
    pack forget .1
  }
}

proc addBindTag {w tag} {
  bindtags $w [concat $tag [bindtags $w]]
}

proc cmdEditNote {id} {
}

proc initGui {} {
  wm title . "Test"
  . config -borderwidth 5

  # Menu frame
  frame .0

  # Search frame
  frame .1
  frame .1.find
  frame .1.results
  frame .1.results.left
  frame .1.results.right
  frame .1.results.right.bnpanel

  entry .1.find.e -width 40
  button .1.find.go -text "Find" -command exit

  listbox .1.results.left.lb -width 40 -height 10
  frame .1.results.right.note
  text .1.results.right.note.text -width 80 -height 20 \
    -yscrollcommand {.1.results.right.note.sb set}
  scrollbar .1.results.right.note.sb -command {.1.results.right.note.text yview}
  button .1.results.right.bnpanel.edit -text "Edit" -command {}
  button .1.results.right.bnpanel.delete -text "Delete" -command {}

  focus .1.find.go

  pack .1.find.e -side left
  pack .1.find.go -side left -padx 5
  pack .1.results.left.lb -side left -fill both -expand true
  pack .1.results.right.note.text -side left -fill both -expand true
  pack .1.results.right.note.sb -side left -fill y
  pack .1.results.right.note -side top -fill both -expand true
  pack .1.results.right.bnpanel.edit -side left
  pack .1.results.right.bnpanel.delete -side left -padx 10
  pack .1.results.right.bnpanel -side top -ipady 5

  pack .1.results.left -side left -fill both -expand true
  pack .1.results.right -side left -fill both -expand true -padx 5
  pack .1.find -side top -fill x -ipady 5
  pack .1.results -side top -fill both -expand true
  pack .1 -side top -fill both -expand true

  # Edit frame
  frame .2
  frame .2.note
  frame .2.bnpanel

  text .2.note.text -width 80 -height 20 -yscrollcommand {.2.note.sb set}
  scrollbar .2.note.sb -command {.2.note.text yview}
  button .2.bnpanel.save -text "Save" -command exit
  button .2.bnpanel.cancel -text "Cancel" -command exit

  pack .2.note.text -side left -fill both -expand true
  pack .2.note.sb -side left -fill y
  pack .2.bnpanel.save -side left
  pack .2.bnpanel.cancel -side left -padx 10

  pack .2.note -side top -fill both -expand true
  pack .2.bnpanel -side top -ipady 5
  pack .2 -side top -fill both -expand true

  activateSection search
}

proc selNote {lb} {
  global nodes

  set iSel [$lb curselection]
  if {$iSel eq ""} {
    return
  }

  set tnote ".1.results.right.note.text"
  $tnote delete 1.0 end


  if {$iSel < [llength $nodes]} {
    set node [lindex $nodes $iSel]
    array set rec $node
    $tnote insert 1.0 $rec(body)
  }
}

proc bindGui {} {
  set lbResults .1.results.left.lb
  set tnote .1.results.right.note.text

  bind $lbResults <<ListboxSelect>> {selNote %W}
  addBindTag $tnote ROText;  # Make $tnote readonly

  # ROText bindtag: Only allow nav keys (arrows, pgup/pgdn/home/end).
  bind ROText <KeyPress> {
    if {[lsearch -exact {Prior Next Home End Up Down Left Right} %K] < 0} {
      break
    }
  }
}

proc renderResults {} {
  global nodes
  set lb .1.results.left.lb

  $lb delete 0 end
  set li 0
  foreach node $nodes {
    array set rec $node
    $lb insert $li $rec(title)

    incr li
  }

  # Select first listbox item.
  if {[$lb size] > 0} {
    $lb selection set 0
  }
  selNote $lb
}

#
# Data section
#

# Given list of body lines
# Extract title (first line) and body (2nd up to last lines)
# Return {title body}
proc parse_body {body_lines} {
  return [list \
    [lindex $body_lines 0] \
    [join [lrange $body_lines 1 end] "\n"] \
  ]
}

proc loadNodes {} {
  set nodes [list]

  set nbdata [glob ~/.nbdata]
  set f [open $nbdata r]

  set fields_pattern {^(\S+)\|(\S+)\|([^|]*)}

  array set rec {}
  set cur_id ""
  set cur_body [list]

  while {[gets $f line] >= 0} {
    # Fields line: <id>|<date>|<tags>
    if {[regexp $fields_pattern $line _ id dt tags]} {
      # Start of new node. Finalize and add previous node.
      if {$cur_id ne ""} {
        set ret [parse_body $cur_body]
        set title [lindex $ret 0]
        set body [lindex $ret 1]

        set rec(title) $title
        set rec(body) $body

        lappend nodes [array get rec]
      }

      # Initialize new node with fields.
      array set rec {}
      set rec(id) $id
      set rec(dt) $dt
      set rec(tags) $tags

      set cur_id $id

      # Initialize new body.
      set cur_body [list]
      continue
    }

    # Add body line
    lappend cur_body $line
  }

  # Add the last node.
  if {$cur_id ne ""} {
    set ret [parse_body $cur_body]
    set title [lindex $ret 0]
    set body [lindex $ret 1]

    set rec(title) $title
    set rec(body) $body

    lappend nodes [array get rec]
  }

  close $f

  return $nodes
}

proc matchBodyLines {bodyLines q} {
  foreach line $bodyLines {
    if {[regexp -nocase -- $q $line]} {
      return 1
    }
  }
  return 0
}

proc findNodes {q} {
  set nodes [list]

  set nbdata [glob ~/.nbdata]
  set f [open $nbdata r]

  set fields_pattern {^(\S+)\|(\S+)\|([^|]*)}

  array set rec {}
  set cur_id ""
  set cur_body [list]

  while {[gets $f line] >= 0} {
    # Fields line: <id>|<date>|<tags>
    if {[regexp $fields_pattern $line _ id dt tags]} {
      # Start of new node. Finalize and add previous node.
      if {$cur_id ne "" && ([matchBodyLines $cur_body $q] || [regexp -nocase -- $q $rec(tags)])} {
        set ret [parse_body $cur_body]
        set title [lindex $ret 0]
        set body [lindex $ret 1]

        set rec(title) $title
        set rec(body) $body

        lappend nodes [array get rec]
      }

      # Initialize new node with fields.
      array set rec {}
      set rec(id) $id
      set rec(dt) $dt
      set rec(tags) $tags

      set cur_id $id

      # Initialize new body.
      set cur_body [list]
      continue
    }

    # Add body line
    lappend cur_body $line
  }

  # Add the last node.
  if {$cur_id ne "" && [matchBodyLines $cur_body $q]} {
    set ret [parse_body $cur_body]
    set title [lindex $ret 0]
    set body [lindex $ret 1]

    set rec(title) $title
    set rec(body) $body

    lappend nodes [array get rec]
  }

  close $f

  return $nodes
}

proc main {} {
  global nodes

  #set nodes [loadNodes]
  set nodes [findNodes "lucy"]
  initGui
  bindGui

  renderResults
}

set nodes [list]
main

