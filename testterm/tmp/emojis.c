#include <time.h>
#include <stdlib.h>

#include "term.h"

typedef struct actor {
    char *s;
    int x,y;
    int is_hit;
} Actor;

void write_actor(struct actor a) {
    term_move_cur(a.x, a.y);
    term_writestr(a.s);
}

// Emojis:
// 👾🛸🚀💾❤️😎🧱

#define NCOLS_INV 10
#define NROWS_INV 4

Actor invs[NROWS_INV][NCOLS_INV];

void init_invs() {
    for (int row=0; row < NROWS_INV; row++) {
        for (int col=0; col < NCOLS_INV; col++) {
            Actor *inv = &invs[row][col];
            //inv->s = "👾";👽
            inv->s = "👽";
            inv->y = row * 2 + 1;
            inv->x = col * 3 + 1;
        }
    }
}

void write_invs() {
    for (int row=0; row < NROWS_INV; row++) {
        for (int col=0; col < NCOLS_INV; col++) {
            Actor inv = invs[row][col];
            write_actor(inv);
        }
    }
}

int rand_n(int n) {
    // Return random number from 0 to n-1.
    return rand() % n;
}

void do_exit() {
    term_restore_settings();
    term_clear_screen();
    term_show_cur();
}

void draw_bricks(int x, int y, int w, int h) {
    term_set_attr(WHITE, BLACK);
    for (int row=y; row < y+h; row++) {
        for (int col=x; col < x+w; col+=1) {
            term_move_cur(col, row);
            term_set_attr(BLACK, WHITE);
            term_writestr(" ");
        }
    }
}

int main() {
    srand(time(NULL));

    term_save_settings();
    atexit(do_exit);

    term_set_raw_mode();
    term_writef("\x1b[=0h");

    int nrows, ncols;
    term_window_size(&nrows, &ncols);

    term_clear_screen();
    term_hide_cur();

    init_invs();
    write_invs();

    int x = 1;
    int y = nrows-1;
    long int seq = 0;

    char k;
    while ((k = term_read_key()) != CTRL_KEY('q')) {
        term_move_cur(1, nrows);
        term_clear_line();
        term_set_attr(BLACK, WHITE);
        term_writef("y: %d  x: %d  seq: %d", y, x, seq);

        seq++;
        if (k == NOKEY) {
            continue;
        }

        if (k == LEFT) {
            x--;
            if (x < 1) {
                x = 1;
            }
        } else if (k == RIGHT) {
            x++;
            if (x > ncols-1) {
                x = ncols-1;
            }
        }

        draw_bricks(x,y-1, ncols, 1);

        term_move_cur(x,y);
        term_clear_line();

        // 0:normal text, 37:white fg, 40:black bg
        term_writestr("😎");
    }

    term_show_cur();
}

