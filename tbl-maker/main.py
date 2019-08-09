import sys

import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk, Pango, Gdk, GLib

import db
import ui
from edit import EditRecWin

class Bag:
    pass

class MainWin(Gtk.Window):
    width = 1000
    height = int(width * 2/3)

    def __init__(self):
        super().__init__(border_width=0, title="TblMaker")
        self.set_size_request(MainWin.width, MainWin.height)

        state = Bag()
        state.tbl = ""
        state.qfind = ""
        state.rec = None
        state.editwins = {}
        self.state = state

        self.setup()


    def setup(self):
        if len(sys.argv) > 1:
            dbfile = sys.argv[1]
            con = db.connect_db(dbfile)
        else:
            con = db.connect_db()
        self.con = con

        self.setup_widgets()

        # Select first table by default
        all_tbls = all_tables(con)
        if len(all_tbls) > 0:
            self.select_table(all_tbls[0])

        # Query all records
        self.run_query(self.state.tbl, "", "")


    def setup_widgets(self):
        self.connect("destroy", Gtk.main_quit)

        # accelerators
        CTRL = Gdk.ModifierType.CONTROL_MASK
        ALT = Gdk.ModifierType.MOD1_MASK
        SHIFT = Gdk.ModifierType.SHIFT_MASK

        widget_accel_group = Gtk.AccelGroup.new()
        self.add_accel_group(widget_accel_group)

        # Content container
        vbox = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=0)
        self.add(vbox)

        # Menu
        menubar = self.create_menubar()
        vbox.pack_start(menubar, False, False, 0)

        # Grid
        grid = Gtk.Grid()
        grid.props.margin = 10
        grid.set_row_spacing(5)
        grid.set_column_spacing(5)
        vbox.pack_start(grid, True, True, 0)

        # Info label
        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=10)
        grid.attach(hbox, 0,1, 10,1)

        info = Gtk.Label("Info")
        info.set_xalign(0)
        hbox.pack_start(info, True, True, 0)

        # Results listbox
        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=10)
        grid.attach(hbox, 0,2, 10,4)

        results_sw = Gtk.ScrolledWindow()
        hbox.pack_start(results_sw, True, True, 0)
        results_sw.set_hexpand(True)
        results_sw.set_vexpand(True)
        results_sw.set_policy(Gtk.PolicyType.AUTOMATIC, Gtk.PolicyType.AUTOMATIC)

        sw_width = int(MainWin.width * 1/3)
        results_sw.set_size_request(sw_width, 0) # set results width

        self.add_results_listbox(results_sw)

        # Rec Preview multiline label
        sw = Gtk.ScrolledWindow()
        hbox.pack_start(sw, True, True, 0)
        sw.set_hexpand(True)
        sw.set_vexpand(True)
        sw.set_policy(Gtk.PolicyType.AUTOMATIC, Gtk.PolicyType.AUTOMATIC)

        frame = Gtk.Frame()
        sw.add(frame)

        preview = Gtk.Label()
        frame.add(preview)
        preview.set_xalign(0)
        preview.set_yalign(0)
        preview.set_line_wrap(True)

        # New, Edit, Del action buttons
        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=5)
        grid.attach(hbox, 6,10, 3,1)

        delete = Gtk.Button(label="Del")
        hbox.pack_end(delete, False, False, 0)

        edit = Gtk.Button(label="Open")
        hbox.pack_end(edit, False, False, 0)

        new = Gtk.Button(label="New")
        hbox.pack_end(new, False, False, 0)

        new.tag = "record_new"
        edit.tag = "record_edit"
        delete.tag = "record_delete"

        new.connect("clicked", self.on_activate)
        edit.connect("clicked", self.on_activate)
        delete.connect("clicked", self.on_activate)

        # No rec selected initially, so disable Edit/Del buttons.
        edit.set_sensitive(False)
        delete.set_sensitive(False)

        ui = Bag()
        ui.menubar      = menubar
        ui.info         = info
        ui.results_sw   = results_sw
        ui.preview      = preview
        ui.new          = new
        ui.edit         = edit
        ui.delete       = delete

        self.ui = ui


    def get_results_lb(self):
        viewport = self.ui.results_sw.get_child()
        results_lb = viewport.get_child()
        return results_lb


    def add_results_listbox(self, results_sw):
        def add_header_sep(w, data):
            sep = Gtk.Separator(orientation=Gtk.Orientation.HORIZONTAL)
            w.set_header(sep)

        results = Gtk.ListBox()
        results_sw.add(results)
        results.set_header_func(add_header_sep)
        results.set_activate_on_single_click(False) # Need double click for row activate
        results.set_can_focus(True)

        results.connect("row-selected", self.on_results_select)
        results.connect("row-activated", self.on_results_activate)

        results_sw.show_all()


    def clear_results(self):
        for child in self.ui.results_sw.get_children():
            self.ui.results_sw.remove(child)

        self.add_results_listbox(self.ui.results_sw)

        # Clear preview pane.
        self.ui.preview.set_markup("")

        # Disable Edit/Del buttons and menuitems.
        self.ui.edit.set_sensitive(False)
        self.ui.delete.set_sensitive(False)
        ui.find_menuitem(self.ui.menubar, "record_edit").set_sensitive(False)
        ui.find_menuitem(self.ui.menubar, "record_delete").set_sensitive(False)

    def on_results_select(self, w, row):
        if row:
            # Remember currently selected rec.
            self.state.rec = row.rec

            # Update preview pane.
            self.ui.preview.set_markup(rec_preview(row.rec))

            # A rec was selected, so enable Edit/Del buttons and menuitems.
            self.ui.edit.set_sensitive(True)
            self.ui.delete.set_sensitive(True)
            ui.find_menuitem(self.ui.menubar, "record_edit").set_sensitive(True)
            ui.find_menuitem(self.ui.menubar, "record_delete").set_sensitive(True)
        else:
            self.state.rec = None

            # Clear preview pane.
            self.ui.preview.set_markup("")

            # No rec selected, so disable Edit/Del buttons and menuitems.
            self.ui.edit.set_sensitive(False)
            self.ui.delete.set_sensitive(False)
            ui.find_menuitem(self.ui.menubar, "record_edit").set_sensitive(False)
            ui.find_menuitem(self.ui.menubar, "record_delete").set_sensitive(False)

    # Results listbox double-click is same as selecting row and clicking 'Edit' button.
    def on_results_activate(self, w, row):
        ui.find_menuitem(self.ui.menubar, "record_edit").activate()
        pass


    def run_query(self, tbl, qfind, qwhere):
        recs = db.read_recs(self.con, tbl, qfind, qwhere)
        print(f"run_query() qfind='{qfind}' #recs={len(recs)}")

        self.clear_results()
        results = self.get_results_lb()
        for rec in recs:
            row = Gtk.ListBoxRow()
            row.rec = rec

            body = ui.escape(rec_row_excerpt(rec))
            lbl = Gtk.Label()
            lbl.set_xalign(0)
            lbl.set_line_wrap(True)
            lbl.set_ellipsize(Pango.EllipsizeMode.END)
            lbl.set_lines(3)
            lbl.set_markup(body)
            row.add(lbl)

            results.add(row)

        results.show_all()

        self.state.qfind = qfind
        self.update_info()


    def run_delete(self, tbl, recid):
        # Delete rec
        db.del_rec(self.con, tbl, recid)
        db.commit(self.con)

        # Close any open edit window for this rec.
        if recid in self.state.editwins:
            childw = self.state.editwins[recid]
            childw.destroy()
            self.edit_wins.pop(recid, None)


    def refresh_results(self, rec, tbl):
        """ Refresh results listbox with changes from rec from table tbl.
            Also restore the last selected row and preview pane.
            If rec is None, it means rec was deleted."""

        # If changed rec is in a different table, ignore it.
        if tbl != self.state.tbl:
            return

        # Update rec in results list if it's there.
        is_row_updated = False
        if rec:
            results = self.get_results_lb()
            for row in results:
                if row.rec['_id'] == rec['_id']:
                    rowlbl = row.get_children()[0]
                    body = ui.escape(rec_row_excerpt(rec))
                    rowlbl.set_markup(body)
                    row.rec = rec

                    is_row_updated = True
                    break

        # rec not in results list, so re-run the search query
        # If it's a new rec, it might show up in the refreshed search.
        # If a rec was deleted (rec == None), it might disappear in the refreshed search.
        if not is_row_updated:
            self.run_query(self.state.tbl, self.state.qfind, "")

            # Restore last selected row
            if self.state.rec:
                sel_recid = self.state.rec['_id']
                results = self.get_results_lb()
                for row in results:
                    if row.rec['_id'] == sel_recid:
                        results.select_row(row)
                        break

        # If edited rec is selected, update preview pane.
        if rec and self.state.rec:
            if rec['_id'] == self.state.rec['_id']:
                self.ui.preview.set_markup(rec_preview(rec))


    def create_menubar(self):
        table_menu = ui.create_table_menu("Table", all_tables(self.con), self.on_activate, self.on_activate_table)

        menubar = ui.create_menubar({
            "Tbl_Maker": ["_About...", "E_xit"],
            "_Table": table_menu,
            "_Record": ["_New...", "_Edit...", "_Delete"],
            "_Search": ["_Text", "_Query", "---", "_Incremental"],
            "T_ools": ["_Tables..."],
        }, self.on_activate)

        # accelerators
        CTRL = Gdk.ModifierType.CONTROL_MASK
        ALT = Gdk.ModifierType.MOD1_MASK
        SHIFT = Gdk.ModifierType.SHIFT_MASK

        accel_group = Gtk.AccelGroup.new()
        self.add_accel_group(accel_group)
        ui.set_menu_accelerators(menubar, accel_group, {
            "tblmaker_exit":        [(ord('q'), CTRL)],
            "record_new":           [(ord('n'), 0)],
            "record_edit":          [(ord('e'), 0), (ord('o'), 0), (Gdk.KEY_Return, 0)],
            "record_delete":        [(ord('x'), 0)],
            "search_text":          [(ord('k'), CTRL)],
            "search_incremental":   [(ord('i'), CTRL | SHIFT)],
        })

        return menubar


    def refresh_tables(self):
        """ Refresh Table menu with latest list of tables from the db. """

        all_tbls = all_tables(self.con)

        # Move each table menuitem from tmp to tables menu
        latest_tables_menu = ui.create_table_menu("Table", all_tbls, self.on_activate, self.on_activate_table)
        ui.replace_menu(ui.find_submenu(self.ui.menubar, "table"), latest_tables_menu)

        # Refresh current selected table
        # If current selected table no longer exists, select the first table from list.
        if self.state.tbl in all_tbls:
            self.select_table(self.state.tbl)
        elif len(all_tbls) > 0:
            self.select_table(all_tbls[0])


    def on_activate(self, w):
        tag = w.tag
        print(f"on_activate(): {tag}")

        if tag == "tblmaker_exit":
            Gtk.main_quit()
        elif tag == "table_new-table":
            dlg = CreateTableDlg(self)
            resp = dlg.run()
            tbl = dlg.table_name()
            dlg.destroy()

            # Create new table.
            if resp == Gtk.ResponseType.OK and tbl != "":
                db.init_table(self.con, tbl)
                db.commit(self.con)

                self.refresh_tables()
                self.select_table(tbl)

                self.run_query(self.state.tbl, "", "")
        elif tag == "record_new":
            # Show Edit Rec window with blank recid.
            editw = EditRecWin(self.con, self.state.tbl, "", self.on_edit_save, self.on_edit_close)
            self.state.editwins[editw.winid] = editw

        elif tag == "record_edit":
            if not self.state.rec:
                return

            # Show Edit Rec window with currect rec selected.
            recid = self.state.rec['_id']
            editw = EditRecWin(self.con, self.state.tbl, recid, self.on_edit_save, self.on_edit_close)
            self.state.editwins[recid] = editw

        elif tag == "record_delete":
            if not self.state.rec:
                return

            dlg = Gtk.MessageDialog(self, 0, Gtk.MessageType.QUESTION, Gtk.ButtonsType.YES_NO, "Delete this rec?")
            dlg.format_secondary_text(rec_row_excerpt(self.state.rec))
            resp = dlg.run()
            dlg.destroy()

            if resp == Gtk.ResponseType.YES:
                recid = self.state.rec['_id']
                self.run_delete(self.state.tbl, recid)

                # Set rec selected to next row's rec.
                # If no next row's rec, set to previous row's rec.
                prev_row = None
                last_prev_row = None
                is_found = False
                for row in self.get_results_lb():
                    if prev_row and self.state.rec['_id'] == prev_row.rec['_id']:
                        self.state.rec = row.rec
                        is_found = True
                        break

                    last_prev_row = prev_row
                    prev_row = row

                if not is_found:
                    if prev_row == None:
                        # No result rows
                        self.state.rec = None
                    elif last_prev_row and self.state.rec['_id'] == prev_row.rec['_id']:
                        # Deleted last result row
                        self.state.rec = last_prev_row.rec
                    else:
                        # Deleted the only result row
                        self.state.rec = None

                # Remove deleted rec from results.
                self.refresh_results(None, self.state.tbl)

        elif tag == "search_text":
            dlg = FindInputDlg(self)
            resp = dlg.run()
            qfind = dlg.qfind()
            dlg.destroy()

            if resp == Gtk.ResponseType.OK and qfind != "":
                self.run_query(self.state.tbl, qfind, "")

    ### on_activate()


    def on_activate_table(self, mi):
        table = mi.tag
        print(f"on_activate_table(): table='{table}'")
        self.select_table(table)
        self.run_query(self.state.tbl, "", "")


    def select_table(self, tbl):
        self.state.tbl = tbl
        self.set_title(f"TblMaker  -  [{tbl}]")
        self.update_info()

        table_menu = ui.find_submenu(self.ui.menubar, "table")
        ui.mark_menuitem(table_menu, tbl, "✓")


    def update_info(self):
        qfind = self.state.qfind or "<all>"
        sinfo = f"<tt><b>{ui.escape(self.state.tbl)}</b></tt>    <tt>search:</tt> <i>{ui.escape(qfind)}</i>"
        self.ui.info.set_markup(sinfo)


    def on_edit_close(self, winid):
        print("on_edit_close()")
        self.state.editwins.pop(winid, None)


    def on_edit_save(self, rec, tbl):
        print("on_edit_save()")
        self.refresh_results(rec, tbl)
        pass

### MainWin


class CreateTableDlg(Gtk.Dialog):
    def __init__(self, parent):
        Gtk.Dialog.__init__(self, "Create Table", parent, Gtk.DialogFlags.MODAL,
            (Gtk.STOCK_OK, Gtk.ResponseType.OK,
            Gtk.STOCK_CANCEL, Gtk.ResponseType.CANCEL))
        self.set_border_width(10)

        content = self.get_content_area()
        content.set_spacing(10)
        lbl = Gtk.Label("New table:")
        lbl.set_xalign(0)
        content.add(lbl)

        self.tbl_entry = Gtk.Entry(activates_default=True, width_chars=20)
        self.tbl_entry.grab_focus()
        content.add(self.tbl_entry)

        ok = self.get_widget_for_response(Gtk.ResponseType.OK)
        ok.props.can_default = True
        ok.grab_default()

        self.show_all()


    def table_name(self):
        tbl = self.tbl_entry.get_text().strip()
        return tbl
### CreateTableDlg


class FindInputDlg(Gtk.Dialog):
    def __init__(self, parent):
        Gtk.Dialog.__init__(self, "Search", parent, Gtk.DialogFlags.MODAL,
            (Gtk.STOCK_OK, Gtk.ResponseType.OK,
            Gtk.STOCK_CANCEL, Gtk.ResponseType.CANCEL))
        self.set_border_width(10)

        content = self.get_content_area()
        content.set_spacing(10)

        lbl = Gtk.Label("Find text:")
        lbl.set_xalign(0)
        content.add(lbl)

        self.find = Gtk.Entry(activates_default=True, width_chars=50)
        self.find.grab_focus()
        content.add(self.find)

        ok = self.get_widget_for_response(Gtk.ResponseType.OK)
        ok.props.can_default = True
        ok.grab_default()

        self.show_all()


    def qfind(self):
        qfind = self.find.get_text().strip()
        return qfind


### Helper functions
def all_tables(con):
    # Get tables list sorted alphabetically with 'data' appearing first.
    all_tbls = list(db.all_tables(con))
    if "data" in all_tbls:
        all_tbls.remove("data")
    all_tbls.sort()
    all_tbls.insert(0, "data")
    return all_tbls


def rec_row_excerpt(rec):
    body = rec.get('_body') or ""
    body_lines = body.split("\n")

    # Return first line in body that has content. Skip over blank lines.
    for line in body_lines:
        if line.strip() != "":
            return line

    return "(empty content)"


def rec_preview(rec):
    body = rec.get('_body') or ""
    body = ui.escape(body)
    return body + "\n" + rec_fields_preview(rec)


def rec_fields_preview(rec):
    # Get list of dot fields in alphabetical order.
    #   dot fields = [all rec fields] - [base fields]
    base_fields = {'_id', '_body'}
    fields = sorted(set(rec.keys()).difference(base_fields))

    # Get longest field name
    max_field_len = 0
    for field in fields:
        if not db.is_valid_field_name(field):
            continue
        if len(field) > max_field_len:
            max_field_len = len(field)

    field_lines = []
    for field in fields:
        if not db.is_valid_field_name(field):
            continue

        k = "{k: <{w}}".format(k=field, w=max_field_len)
        v = ui.escape(str(rec[field]))

        line = f"<span weight='bold' font-family='monospace'>{k}</span> : {v}"
        field_lines.append(line)

    return "\n".join(field_lines)


if __name__ == "__main__":
    w = MainWin()
    w.show_all()
    Gtk.main()

