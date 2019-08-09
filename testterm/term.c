#include <stdio.h>
#include <stdarg.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
#include <termios.h>
#include <ctype.h>
#include <sys/ioctl.h>

#include "term.h"

void term_fatal(char *s) {
    write(1, "\x1b[2J", 4);
    write(1, "\x1b[H", 3);
    perror(s);
    exit(1);
}

struct termios org_termios;
void term_save_settings() {
    tcgetattr(0, &org_termios);
}
void term_restore_settings() {
    tcsetattr(0, TCSAFLUSH, &org_termios);
}

void term_set_normal_mode() {
    struct termios t;
    tcgetattr(0, &t);
    t.c_iflag |= ~(ICRNL | IXON);
    t.c_oflag |= ~(OPOST);
    t.c_lflag |= ~(ECHO | ICANON | IEXTEN | ISIG);
    tcsetattr(0, TCSAFLUSH, &t);
}

void term_set_raw_mode() {
    struct termios t;
    tcgetattr(0, &t);

    t.c_iflag &= ~(ICRNL | IXON | BRKINT | INPCK | ISTRIP);
    t.c_oflag &= ~(OPOST);
    t.c_cflag |= ~(CS8);
    t.c_lflag &= ~(ECHO | ICANON | IEXTEN | ISIG);

    // Minimum number of bytes read before read() returns.
    t.c_cc[VMIN] = 0;

    // Maximum time in tenths of a second before read() returns.
    t.c_cc[VTIME] = 1;

    tcsetattr(0, TCSAFLUSH, &t);
}

void term_window_size(int *rows, int *cols) {
    *rows = 0;
    *cols = 0;

    struct winsize ws;
    if (ioctl(1, TIOCGWINSZ, &ws) == -1) {
        return;
    }
    *rows = ws.ws_row;
    *cols = ws.ws_col;
}

void term_set_attr(int fg, int bg) {
    term_writestr("\x1b[0;");
    if (fg != -1) {
        term_writef("%d;", fg + 30);
    } else {
        term_writestr(";");
    }

    if (bg != -1) {
        term_writef("%dm", bg + 40);
    } else {
        term_writestr("m");
    }
}

void term_writech(char ch) {
    write(1, &ch, 1);
}

void term_writestr(char *s) {
    write(1, s, strlen(s));
}

void term_writef(const char *format, ...) {
    va_list args;
    va_start(args, format);
    vprintf(format, args);
    fflush(stdout);
    va_end(args);
}

void term_clear_screen() {
    // 0:normal text, 37:white fg, 40:black bg
    term_writestr("\x1b[0;37;40m");

    // Clear entire screen.
    term_writestr("\x1b[2J");

    // Move cursor to home 1,1 position.
    term_writestr("\x1b[H");
}

void term_clear_line() {
    // 0:normal text, 37:white fg, 40: black bg
    term_writestr("\x1b[0;37;40m");
    term_writestr("\x1b[2K");
}

void term_print_key(char k) {
    switch (k) {
    case NOKEY:
        term_writestr("(NOKEY)");
        break;
    case LEFT:
        term_writestr("(LEFT)");
        break;
    case RIGHT:
        term_writestr("(RIGHT)");
        break;
    case UP:
        term_writestr("(UP)");
        break;
    case DOWN:
        term_writestr("(DOWN)");
        break;
    case HOME:
        term_writestr("(HOME)");
        break;
    case END:
        term_writestr("(END)");
        break;
    case PGUP:
        term_writestr("(PGUP)");
        break;
    case PGDN:
        term_writestr("(PGDN)");
        break;
    default:
        if (IS_CTRL_KEY(k)) {
            char ctrl_ch = GET_CTRL_CHAR(k);
            term_writestr("CTRL-");
            write(1, &ctrl_ch, 1);
        } else if (iscntrl(k)) {
            term_writestr("(control)");
        } else {
            write(1, &k, 1);
        }
    }
}

char term_read_byte() {
    char c = NOKEY;
    int nread = read(0, &c, 1);
    if (nread == -1) {
        term_fatal("term_read_byte() error");
    }
    return c;
}

char term_read_key() {
    char c;

    // If no new byte read, return that no key is available.
    c = term_read_byte();
    if (c == NOKEY) {
        return NOKEY;
    }

    // Escape sequences start with '\x1b', '['.

    // If not an escape sequence, return the byte read.
    if (c != '\x1b') {
        return c;
    }
    if (term_read_byte() != '[') {
        return '\x1b';
    }

    // Escape sequence read. The next two bytes determine what key was entered:
    // "A" = Up arrow
    // "B" = Down arrow
    // "C" = Right arrow
    // "D" = Left arrow
    //
    // "1~" = Home
    // "4~" = End
    // "5~" = PgUp
    // "6~" = PgDn

    c = term_read_byte();
    if (c == 'A') {
        return UP;
    } else if (c == 'B') {
        return DOWN;
    } else if (c == 'C') {
        return RIGHT;
    } else if (c == 'D') {
        return LEFT;
    } else if (c == '1' || c == '4' || c == '5' || c == '6') {
        if (term_read_byte() != '~') {
            return c;
        }
        if (c == '1') {
            return HOME;
        } else if (c == '4') {
            return END;
        } else if (c == '5') {
            return PGUP;
        } else if (c == '6') {
            return PGDN;
        }
    }

    return c;
}

void term_move_cur(int x, int y) {
    term_writef("\x1b[%d;%df", y, x);
}

void term_hide_cur() {
    term_writef("\x1b[?25l");
}
void term_show_cur() {
    term_writef("\x1b[?25h");
}

