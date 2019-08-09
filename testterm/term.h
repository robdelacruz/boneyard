enum {
    NOKEY = '\x80',
    LEFT,
    RIGHT,
    UP,
    DOWN,
    HOME,
    END,
    PGUP,
    PGDN
};

enum {
    BLACK = 0,
    RED,
    GREEN,
    YELLOW,
    BLUE,
    MAGENTA,
    CYAN,
    WHITE
};

#define CTRL_KEY(k) ((k) & 0x1f)
#define IS_CTRL_KEY(k) (k >= 1 && k <= 26)
#define GET_CTRL_CHAR(k) ((k & 0x1f) + 'a' - 1)

void term_fatal(char *s);
void term_save_settings();
void term_restore_settings();
void term_set_normal_mode();
void term_set_raw_mode();
void term_window_size(int *rows, int *cols);
void term_set_attr(int fg, int bg);
void term_writech(char ch);
void term_writestr(char *s);
void term_writef(const char *format, ...);
void term_clear_screen();
void term_clear_line();
void term_print_key(char k);
char term_read_byte();
char term_read_key();
void term_move_cur(int x, int y);
void term_hide_cur();
void term_show_cur();

