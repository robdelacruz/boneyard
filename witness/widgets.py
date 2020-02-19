import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk, Pango, Gdk, GLib

import db
import ui
import conv

def set_widget_visible(visible, *widgets):
    for w in widgets:
        w.set_visible(visible)


class EntryForm:
    def __init__(self, con, dt=None, entrytype=None, on_save=None, on_cancel=None):
        self._con = con
        if not dt:
            dt = conv.today_isodt()
        self._dt = dt

        if not entrytype:
            entrytype = db.EntryType.LOG

        frame = Gtk.Frame()
        grid = Gtk.Grid()
        grid.props.margin = 2
        grid.set_row_spacing(2)
        grid.set_column_spacing(0)
        frame.add(grid)

        heading_lbl = Gtk.Label("[New Entry]")
        heading_lbl.set_xalign(0)
        grid.attach(heading_lbl, 0,0, 1,1)

        type_cb = Gtk.ComboBoxText()
        type_cb.append_text("log")
        type_cb.append_text("note")
        type_cb.append_text("event")
        type_cb.append_text("task")
        type_cb.set_active(entrytype)
        type_cb.set_halign(Gtk.Align.END)
        grid.attach(type_cb, 4,0, 1,1)

        y = 1

        body_tv = Gtk.TextView()
        body_tv.set_wrap_mode(Gtk.WrapMode.WORD)
        sw = ui.scrollwindow(body_tv)
        grid.attach(sw, 0,y, 5,1)
        y += 1

        startdt_chk = Gtk.CheckButton("Start on")
        startdt_entry = Gtk.Entry()
        grid.attach(startdt_chk, 0,y, 1,1)
        grid.attach(startdt_entry, 1,y, 1,1)
        y += 1

        enddt_chk = Gtk.CheckButton("End on")
        enddt_entry = Gtk.Entry()
        grid.attach(enddt_chk, 0,y, 1,1)
        grid.attach(enddt_entry, 1,y, 1,1)
        y += 1

        status_lbl = Gtk.Label("Status")
        status_lbl.set_xalign(0)
        status_cb = Gtk.ComboBoxText()
        status_cb.append_text("Open")
        status_cb.append_text("Closed")
        status_cb.set_active(0)
        grid.attach(status_lbl, 0,y, 1,1)
        grid.attach(status_cb, 1,y, 1,1)
        y += 1

        lbl = Gtk.Label("Assign to topic")
        lbl.set_xalign(0)
        topic_cb = Gtk.ComboBoxText.new_with_entry()
        topic_cb.append_text("Unassigned")
        topic_cb.append_text("Cats")
        topic_cb.append_text("The Blog")
        topic_cb.set_active(0)
        grid.attach(lbl, 0,y, 1,1)
        grid.attach(topic_cb, 1,y, 1,1)
        y += 1

        save = Gtk.Button("Save")
        cancel = Gtk.Button("Cancel")
        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=10)
        hbox.pack_start(save, False, False, 0)
        hbox.pack_start(cancel, False, False, 0)
        grid.attach(hbox, 0,y, 5,1)
        y += 1

        def show_chk_entry(chk, entry):
            if chk.get_active():
                entry.set_sensitive(True)
            else:
                entry.set_sensitive(False)

        show_chk_entry(startdt_chk, startdt_entry)
        show_chk_entry(enddt_chk, enddt_entry)

        type_cb.connect("changed", self.show_visible_widgets)
        startdt_chk.connect("toggled", show_chk_entry, startdt_entry) 
        enddt_chk.connect("toggled", show_chk_entry, enddt_entry) 

        def on_save_clicked(w):
            entry = self.get_entry()
            db.add_entry(self._con, entry)
            db.commit(self._con)
            self.clear()

            if self.on_save:
                self.on_save(entry)
        save.connect("clicked", on_save_clicked)

        def on_cancel_clicked(w):
            if self.on_cancel:
                self.on_cancel()
        cancel.connect("clicked", on_cancel_clicked)

        self._frame = frame
        self._grid = grid

        self._entry_id = None
        self._type_cb = type_cb
        self._heading_lbl = heading_lbl
        self._body_tv = body_tv
        self._startdt_chk = startdt_chk
        self._enddt_chk = enddt_chk
        self._startdt_entry = startdt_entry
        self._enddt_entry = enddt_entry
        self._status_lbl = status_lbl
        self._status_cb = status_cb
        self._topic_cb = topic_cb
        self.savebtn = save
        self.cancelbtn = cancel
        self.on_save = on_save
        self.on_cancel = on_cancel

        self.refresh_heading()


    def widget(self):
        return self._frame


    def clear(self):
        buf = self._body_tv.get_buffer()
        buf.set_text("")
        self._startdt_chk.set_active(False)
        self._enddt_chk.set_active(False)
        self._startdt_entry.set_text("")
        self._enddt_entry.set_text("")
        self._status_cb.set_active(0)
        self._topic_cb.set_active(0)
        self.show_visible_widgets()
        self._entry_id = None


    def set_date(self, dt):
        if dt == self._dt:
            return

        self._dt = dt
        self.refresh_heading()


    def refresh_heading(self):
        longdate = conv.isodt_to_longfmt(self._dt)
        self._heading_lbl.set_text(f"{longdate}  [New Entry]")


    def show_visible_widgets(self, *args):
        entrytype = self._type_cb.get_active_text()

        show_dates = False
        if entrytype == "event" or entrytype == "task":
            show_dates = True
        set_widget_visible(show_dates, self._startdt_chk, self._startdt_entry, self._enddt_chk, self._enddt_entry)

        if entrytype == "task":
            self._startdt_chk.set_label("Alert on")
            self._enddt_chk.set_label("Due on")
            set_widget_visible(True, self._status_lbl, self._status_cb)
        else:
            self._startdt_chk.set_label("Start on")
            self._enddt_chk.set_label("End on")
            set_widget_visible(False, self._status_lbl, self._status_cb)


    def get_entry(self):
        entry = {}
        entry["id"] = self._entry_id

        # type
        sel = self._type_cb.get_active_text()
        entrytype = 0
        if sel == "log":
            entrytype = db.EntryType.LOG
        if sel == "note":
            entrytype = db.EntryType.NOTE
        elif sel == "event":
            entrytype = db.EntryType.EVENT
        elif sel == "task":
            entrytype = db.EntryType.TASK
        entry["type"] = entrytype

        entry["createdt"] = self._dt
        entry["startdt"] = None
        entry["enddt"] = None
        if entrytype == db.EntryType.EVENT or entrytype == db.EntryType.TASK:
            if self._startdt_chk.get_active() and self._startdt_entry.get_text().strip() != "":
                entry["startdt"] = self._startdt_entry.get_text()
            if self._enddt_chk.get_active() and self._enddt_entry.get_text().strip() != "":
                entry["enddt"] = self._enddt_entry.get_text()

        entry["status"] = db.StatusType.OPEN
        if entrytype == db.EntryType.TASK:
            entry["status"] = self._status_cb.get_active()

        buf = self._body_tv.get_buffer()
        s = buf.get_start_iter()
        e = buf.get_end_iter()
        entry["body"] = buf.get_text(s, e, True).strip()

        # self._topic_cb.get_active_text()
        entry["topic_id"] = 0
        return entry


    def load_entry(self, entry_id):
        self.clear()

        entry = db.read_entry(self._con, entry_id)
        if not entry:
            return db.EntryType.LOG

        self._entry_id = entry_id

        entrytype = entry.get("type", db.EntryType.LOG)
        body = entry.get("body", "")
        createdt = entry.get("createdt", "")
        startdt = entry.get("startdt", "")
        enddt = entry.get("enddt", "")
        status = entry.get("status", db.StatusType.OPEN)
        topic_id = entry.get("topic_id", 0)

        self.set_date(createdt)
        self._type_cb.set_active(entrytype)
        self._startdt_entry.set_text(startdt)
        self._enddt_entry.set_text(enddt)
        self._status_cb.set_active(status)
        self._topic_cb.set_active(topic_id)
        buf = self._body_tv.get_buffer()
        buf.set_text(body)

        if startdt:
            self._startdt_chk.set_active(True)
        else:
            self._startdt_chk.set_active(False)
        if enddt:
            self._enddt_chk.set_active(True)
        else:
            self._enddt_chk.set_active(False)

        self.show_visible_widgets()

        return entry.get("type", db.EntryType.LOG)


class EditEntryWin(Gtk.Window):
    width = 300
    height = int(width * 3/2)

    def __init__(self, con, **kwargs):
        dt = kwargs.get("dt")
        on_save = kwargs.get("on_save")
        entry_id = kwargs.get("entry_id")
        entrytype = kwargs.get("entrytype")

        super().__init__(border_width=0, title="Edit Entry")
        self.set_size_request(EditEntryWin.width, EditEntryWin.height)

        def _on_save(entry):
            if on_save:
                on_save(entry)
            self.destroy()
        entryform = EntryForm(con, dt, entrytype, _on_save, self.destroy)

        # accelerators
        CTRL = Gdk.ModifierType.CONTROL_MASK
        ALT = Gdk.ModifierType.MOD1_MASK
        SHIFT = Gdk.ModifierType.SHIFT_MASK

        accel_group = Gtk.AccelGroup.new()
        self.add_accel_group(accel_group)
        entryform.savebtn.add_accelerator("clicked", accel_group, ord('s'), CTRL, Gtk.AccelFlags.VISIBLE)
        entryform.cancelbtn.add_accelerator("clicked", accel_group, Gdk.KEY_Escape, 0, Gtk.AccelFlags.VISIBLE)

        self.add(entryform.widget())
        self.show_all()

        if entry_id:
            entrytype = entryform.load_entry(entry_id)

            title = ""
            if entrytype == db.EntryType.LOG:
                title = "Edit Log"
            elif entrytype == db.EntryType.NOTE:
                title = "Edit Note"
            elif entrytype == db.EntryType.EVENT:
                title = "Edit Event"
            elif entrytype == db.EntryType.TASK:
                title = "Edit Task"
            self.set_title(title)
        else:
            entryform.show_visible_widgets()

            title = ""
            if entrytype == db.EntryType.LOG:
                title = "New Log"
            elif entrytype == db.EntryType.NOTE:
                title = "New Note"
            elif entrytype == db.EntryType.EVENT:
                title = "New Event"
            elif entrytype == db.EntryType.TASK:
                title = "New Task"
            self.set_title(title)


class EntryPreview:
    def __init__(self, entry={}):
        lbl = Gtk.Label()
        lbl.set_hexpand(True)
        lbl.set_vexpand(False)
        lbl.set_xalign(0)
        lbl.set_yalign(0)
        lbl.set_line_wrap(True)
        sw = ui.scrollwindow(lbl)

        frame = ui.frame(sw, "")

        self.frame = frame
        self.lbl = lbl
        self.entry = entry

        self.refresh()


    def widget(self):
        return self.frame


    def refresh(self):
        body = self.entry.get("body", "")
        entrytype = self.entry.get("type", -1)

        self.lbl.set_markup(body)

        heading = ""
        if entrytype == db.EntryType.LOG:
            heading = "Log Entry"
        elif entrytype == db.EntryType.NOTE:
            heading = "Note"
        elif entrytype == db.EntryType.EVENT:
            heading = "Event"
        elif entrytype == db.EntryType.TASK:
            heading = "Task"
        self.frame.heading_lbl.set_text(heading)
        self.frame.show_all()


    def load_entry(self, entry):
        self.entry = entry
        self.refresh()


def icon_image(icon_name):
    theme = Gtk.IconTheme.get_default()
    icon = theme.load_icon(icon_name, -1, Gtk.IconLookupFlags.FORCE_SIZE)
    img = Gtk.Image.new_from_pixbuf(icon)
    return img


def event_daterange_text(entry):
    startdt = entry.get("startdt")
    enddt = entry.get("enddt")

    if startdt == enddt:
        return conv.isodt_to_shortfmt(enddt)
    else:
        return f"{conv.isodt_to_shortfmt(startdt)} to {conv.isodt_to_shortfmt(startdt)}"


def task_date_text(entry):
    enddt = entry.get("enddt")

    if not enddt:
        return ""
    else:
        return f"Due on {conv.isodt_to_shortfmt(enddt)}"


def journal_widget_entry(entry):
    icon_img = None

    # Limit listbox text to the first paragraph
    entry_text = entry.get("body", "")
    entry_text = entry_text.strip().split("\n")[0]

    entry_footer = ""
    entrytype = entry.get("type")
    if entrytype == db.EntryType.LOG:
        icon_img = icon_image("list-remove")
        #icon_img = icon_image("user-status-pending")
        #icon_img = icon_image("pan-end")
        #icon_img = icon_image("document-edit")
    elif entrytype == db.EntryType.NOTE:
        #icon_img = icon_image("document-edit")
        icon_img = icon_image("text-x-generic")
    elif entrytype == db.EntryType.EVENT:
        #icon_img = icon_image("appointment-soon")
        icon_img = icon_image("x-office-calendar")
        entry_footer = event_daterange_text(entry)
    elif entrytype == db.EntryType.TASK:
        icon_img = icon_image("task-due")
        entry_footer = task_date_text(entry)
    else:
        icon_img = icon_image("list-remove")

    entry_text = ui.escape(entry_text)
    entry_footer = ui.escape(entry_footer)

    lbl = Gtk.Label()
    lbl.set_hexpand(True)
    lbl.set_xalign(0)
    lbl.set_valign(Gtk.Align.CENTER)
    lbl.set_line_wrap(True)
    lbl.set_ellipsize(Pango.EllipsizeMode.END)
    lbl.set_lines(3)

    lbl.set_max_width_chars(50)
    lbl.set_margin_top(4)
    lbl.set_margin_bottom(4)
    lbl.set_margin_left(4)
    lbl.set_markup(entry_text)

    vbox = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=0)
    vbox.add(lbl)
    if entry_footer:
        lbl_footer = Gtk.Label()
        lbl_footer.set_xalign(0)
        lbl_footer.set_margin_left(4)
        lbl_footer.set_markup(f"<span fgcolor='darkgrey'>{entry_footer}</span>")
        vbox.add(lbl_footer)

    hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=10)
    hbox.add(icon_img)
    hbox.add(vbox)

    return hbox


class JournalView:
    def __init__(self, con, **opts):
        """
        opts:
            dt              = iso date string of date param to pass to read_fn function
            read_fn         = function returning entries
            heading         = text heading to appear
            on_sel_entry    = function to call when entry item selected
            show_preview    = bool whether to show preview pane
        """
        dt = opts.get("dt", conv.today_isodt())
        read_fn = opts.get("read_fn", db.read_today_all_entries)
        heading = opts.get("heading", "")
        on_sel_entry = opts.get("on_sel_entry")
        show_preview = opts.get("show_preview", False)

        entries_lb = Gtk.ListBox()

        def on_sel(lb, row):
            if not row:
                return
            entry_id = row.entry_id
            entry = db.read_entry(self._con, entry_id)
            if not entry:
                return
            if self._preview:
                self._preview.load_entry(entry)
            if on_sel_entry:
                on_sel_entry(entry)
        entries_lb.connect("row-selected", on_sel)
        sw = ui.scrollwindow(entries_lb)

        preview = None
        if show_preview:
            preview = EntryPreview()
            vpane = Gtk.VPaned()
            vpane.add1(sw)
            vpane.add2(preview.widget())
            vpane.set_position(400)
            frame = ui.frame(vpane, "")
        else:
            frame = ui.frame(sw, "")

        self._con = con
        self._dt = dt
        self._frame = frame
        self._heading_lbl = frame.heading_lbl
        self._entries_lb = entries_lb
        self._preview = preview
        self._read_fn = read_fn
        self._heading = heading

        self.refresh()


    def widget(self):
        return self._frame


    def refresh(self):
        ui.clear_lb(self._entries_lb)

        entries = self._read_fn(self._con, self._dt)
        for entry in entries:
            row = Gtk.ListBoxRow()
            row.entry_id = entry.get("id", "")
            lbl = journal_widget_entry(entry)
            row.add(lbl)
            self._entries_lb.add(row)
        self._entries_lb.show_all()

        heading = self._heading
        if heading == None:
            heading = conv.isodt_to_longfmt(self._dt)
        self._heading_lbl.set_text(heading)


    def clear_sel_highlight(self):
        self._entries_lb.unselect_all()


    def set_date(self, dt):
        self._dt = dt
        self.refresh()


    def get_selected_entry_id(self):
        row = self._entries_lb.get_selected_row()
        if row:
            return row.entry_id

