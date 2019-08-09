#!/usr/bin/env tclsh

proc joinLines {lines} {
  return [join $lines "\n"]
}

proc nbdata_ParseFile {datafile {filterCmd ""}} {
  # %%<node id>
  # Ex. %%12345
  set idPattern {^%%(\S+)\s*$}

  # .field1 <val>
  # Ex.
  # .author robtwister
  # .date 2018-08-01
  # .tags cats, national cat day, lucy    
  set dotFieldPattern {^\.(\S+)\s+(.*)\s*$}

  set node [dict create]
  set nodeID ""
  set nodeBodyLines [list]

  set nodes [list]
  set f [open $datafile r]
  while {[gets $f line] >= 0} {
    if {[regexp $idPattern $line _ id]} {
      # ID line - start new node
      if {$nodeID ne ""} {
        # Add previous node to list.
        dict set node body [joinLines $nodeBodyLines]
        puts "nbdata_ParseFile(): filterCmd='$filterCmd'"
        if {$filterCmd eq "" || [eval $filterCmd node]} {
          lappend nodes $node
        }
      }

      set node [dict create]
      dict set node id $id

      set nodeID $id
      set nodeBodyLines [list]
    } elseif {[regexp $dotFieldPattern $line _ field val]} {
      # Dot field line
      dict set node $field $val
    } else {
      # Body line
      lappend nodeBodyLines $line
    }
  }

  # Add last node.
  if {$nodeID ne ""} {
    # Add previous node to list.
    dict set node body [joinLines $nodeBodyLines]
    if {$filterCmd eq "" || [eval $filterCmd node]} {
      lappend nodes $node
    }
  }

  close $f
  return $nodes
}

proc nbdata_FindNodes {nodesVar q} {
  upvar $nodesVar nodes

  set matchedNodes [list]
  set found 0

  foreach node $nodes {
    if {[matchRec $q $node]} {
      lappend matchedNodes $node
    }
  }

  return $matchedNodes
}

proc matchRec {q nodeVar} {
  upvar $nodeVar node

  dict for {k v} $node {
    if {$k eq "date" || $k eq "id"} {
      continue
    }

    if {$k eq "body"} {
      set bodyLines [split $v "\n"]
      foreach bodyLine $bodyLines {
        if {[regexp -nocase -- $q $bodyLine]} {
          return 1
        }
      }
    } else {
      if {[regexp -nocase -- $q $v]} {
        return 1
      }
    }
  }

  return 0
}

set q ""
if {[llength $argv] > 0} {
  set q [lindex $argv]
}

set nbdataFile [glob ~/.nbdata2]
set nodes [nbdata_ParseFile $nbdataFile "matchRec {$q}"]
#set nodes [nbdata_ParseFile $nbdataFile]

#set nodes [nbdata_FindNodes nodes "abc"]

foreach node $nodes {
  puts "---------------"
  dict for {k v} $node {
    puts "$k: $v"
  }
}

