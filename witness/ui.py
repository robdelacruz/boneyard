import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk, GLib

def escape(s):
    return GLib.markup_escape_text(s)

### Menu functions
def fmt_tag(lbl):
    # 'New...' --> 'new'
    # '_Edit' --> 'edit'
    # 'To Do'  --> 'to-do'
    lbl = lbl.replace(".", "")
    lbl = lbl.replace("_", "")
    lbl = lbl.strip()
    lbl = lbl.replace(" ", "-")
    return lbl.lower()


def clear_lb(lb):
    lb.foreach(lambda row: lb.remove(row))


def create_menu(top_lbl, menu_items, on_activate=None):
    """ create_menu("File", ["New...", "Edit", "Exit"], on_activate)
        Creates a new dropdown with the menu items specified.

        <top_lbl> is used to tag the menu item with a unique identifier for the
        on_activate handler."""
        
    menu = Gtk.Menu.new()
    menu.tag = fmt_tag(top_lbl)

    for lbl in menu_items:
        if lbl == "---":
            mi = Gtk.SeparatorMenuItem()
            mi.tag = ""
            menu.append(mi)
            continue

        mi = Gtk.MenuItem.new_with_mnemonic(lbl)
        mi.set_use_underline(True)

        # Add a tag so menuitem can be identified from activate handler.
        # format: <parent tag>_<menuitem_tag>
        # Ex. "table_new-table"
        mi.tag = f"{fmt_tag(top_lbl)}_{fmt_tag(lbl)}"

        if on_activate:
            mi.connect("activate", on_activate)
        menu.append(mi)

    return menu


def create_table_menu(top_lbl, tables, on_activate=None, on_activate_table=None):
    """ Create a table selector menu with a 'New Table...' menu item.
        Similar to create_menu()."""
        
    menu = Gtk.Menu.new()
    menu.tag = fmt_tag(top_lbl)

    for table in tables:
        mi = Gtk.MenuItem.new_with_label(table)
        mi.tag = table
        if on_activate_table:
            mi.connect("activate", on_activate_table)
        menu.append(mi)

    mi = Gtk.SeparatorMenuItem()
    mi.tag = ""
    menu.append(mi)

    lbl = "_New Table..."
    mi = Gtk.MenuItem.new_with_mnemonic(lbl)
    mi.tag = f"{fmt_tag(top_lbl)}_{fmt_tag(lbl)}"
    if on_activate:
        mi.connect("activate", on_activate)
    menu.append(mi)

    return menu


def create_menubar(submenus, on_activate=None):
    """ Creates a menubar with submenus.
        Sample usage:
        create_menubar({
            "_File": ["_New", "_Edit", "---", "E_xit"],
            "_Help": ["_Contents", "_About"]
        })"""

    menubar = Gtk.MenuBar.new()

    for top_label, menuitems in submenus.items():
        top_mi = Gtk.MenuItem.new_with_mnemonic(top_label)
        top_mi.set_use_underline(True)
        top_mi.tag = fmt_tag(top_label)

        submenu = None
        if isinstance(menuitems, list):
            submenu = create_menu(top_label, menuitems, on_activate)
        elif isinstance(menuitems, Gtk.Menu):
            submenu = menuitems
        else:
            raise Exception("create_menubar(): pass either list of menuitem labels or list of Gtk.Menu's")

        top_mi.set_submenu(submenu)
        menubar.append(top_mi)

    return menubar


def print_menu(menu, level=0):
    for mi in menu.get_children():
        if type(mi) is not Gtk.MenuItem:
            continue

        indent = "\t" * level
        print(f"{indent}{mi.get_label()}  tag: {mi.tag}")

        submenu = mi.get_submenu()
        if submenu:
            print_menu(submenu, level+1)


def find_menuitem(menubar, tag):
    for mi in menubar.get_children():
        if type(mi) is not Gtk.MenuItem:
            continue

        if mi.tag == tag:
            return mi

        submenu = mi.get_submenu()
        if submenu:
            found_mi = find_menuitem(submenu, tag)
            if found_mi:
                return found_mi

    return None


def find_submenu(menubar, tag):
    for mi in menubar.get_children():
        if type(mi) is not Gtk.MenuItem:
            continue

        submenu = mi.get_submenu()
        if submenu.tag == tag:
            return submenu

    return None


def clear_menu(menu):
    for mi in menu.get_children():
        menu.remove(mi)


def replace_menu(dest_menu, src_menu):
    clear_menu(dest_menu)
    for mi in src_menu.get_children():
        src_menu.remove(mi)
        dest_menu.append(mi)


def set_menu_accelerators(menu, accel_group, accel_map):
    if type(menu) is not Gtk.MenuBar:
        menu.set_accel_group(accel_group)

    for mi in menu.get_children():
        if type(mi) is not Gtk.MenuItem:
            continue

        if mi.tag in accel_map:
            for (key, mods) in accel_map[mi.tag]:
                mi.add_accelerator("activate", accel_group, key, mods, Gtk.AccelFlags.VISIBLE)

        submenu = mi.get_submenu()
        if submenu:
            set_menu_accelerators(submenu, accel_group, accel_map)


def mark_menuitem(menu, tag, mark_text):
    mark_text = f" {mark_text}"

    # Remove mark by resetting all menu items
    for mi in menu.get_children():
        # Stop when menu separator (empty tag) reached
        if mi.tag == "":
            break

        lbl = mi.get_children()[0]
        lbl_text = lbl.get_text()

        # Remove mark text when found.
        if lbl_text.endswith(mark_text):
            lbl.set_markup(escape(lbl_text[:-len(mark_text)]))

    # Make selected table menu item bold and with check mark
    for mi in menu.get_children():
        # Stop when menu separator reached
        if mi.tag == "":
            break
        if mi.tag == tag:
            schk = " ✓"
            lbl = mi.get_children()[0]
            lbl_text = lbl.get_text()

            lbl.set_markup(f"<b>{escape(lbl_text)}{mark_text}</b>")
            menu.show_all()
            break


def heading_vbox(widget, heading):
    vbox = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=5)
    lbl = Gtk.Label(heading)
    vbox.pack_start(lbl, False, False, 0)
    vbox.pack_start(widget, True, True, 0)

    vbox.heading_lbl = lbl
    return vbox


def frame(widget, heading=None):
    frame = Gtk.Frame()

    w = widget
    if heading != None:
        w = heading_vbox(widget, heading)
        frame.heading_lbl = w.heading_lbl

    frame.add(w)
    return frame


def scrollwindow(widget):
    sw = Gtk.ScrolledWindow()
    sw.set_hexpand(True)
    sw.set_vexpand(True)
    sw.set_policy(Gtk.PolicyType.AUTOMATIC, Gtk.PolicyType.AUTOMATIC) 
    sw.add(widget)
    return sw


