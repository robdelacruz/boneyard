#include <time.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

#include "term.h"
#include "gapbuf.h"

void write_buf_to_screen(gapbuf_t* g);

void do_exit() {
    term_restore_settings();
    term_clear_screen();
    term_show_cur();
}

int _win_rows, _win_cols;

int main() {
    term_window_size(&_win_rows, &_win_cols);
    term_save_settings();
    atexit(do_exit);
    term_set_raw_mode();

    term_clear_screen();
    term_hide_cur();

    gapbuf_t* g = gapbuf_new();
    char k;

    while ((k = term_read_key()) != CTRL_KEY('q')) {
        if (k == NOKEY) continue;

        if (isprint(k)) {
            char s[2];
            s[0] = k;
            s[1] = '\0';
            gapbuf_insert_text(g, s);
        } else if (k == 13) {
            gapbuf_insert_text(g, "\n");
        } else if (k == LEFT) {
            gapbuf_shift_gap(g, -1);
        } else if (k == RIGHT) {
            gapbuf_shift_gap(g, 1);
        } else if (k == UP) {
            gapbuf_shift_row(g, -1);
        } else if (k == DOWN) {
            gapbuf_shift_row(g, 1);
        }

        write_buf_to_screen(g);

        term_set_attr(BLACK, WHITE);
        term_move_cur(0,10);
        pos_t pos = gapbuf_get_pos(g);
        term_writef("      (%d,%d)", pos.row, pos.col);
    }

    term_show_cur();
}

void write_buf_to_screen(gapbuf_t* g) {
    term_clear_screen();

    char *text = gapbuf_text(g);
    term_writestr(text);

    free(text);
}



