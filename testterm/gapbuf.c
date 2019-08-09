#include <stdlib.h>
#include <stdio.h>
#include <string.h>

#include "gapbuf.h"

#define GAP_LEN(g)      (g->gap_end - g->gap_start + 1)
#define PRE_LEN(g)      (g->gap_start)
#define POST_LEN(g)     (g->bytes_len - g->gap_end - 1)
#define PRE_START(g)    (0)
#define PRE_END(g)      (g->gap_start-1)
#define POST_START(g)   (g->gap_end+1)
#define POST_END(g)     (g->bytes_len-1)

gapbuf_t* gapbuf_new() {
    int buf_initial_size = 10;

    gapbuf_t* g = malloc(sizeof(gapbuf_t));
    g->bytes = malloc(buf_initial_size);
    memset(g->bytes, 0, buf_initial_size);

    g->bytes_len = buf_initial_size;
    g->gap_start = 0;
    g->gap_end = buf_initial_size-1;

    return g;
}

void gapbuf_free(gapbuf_t* g) {
    free(g->bytes);
    free(g);
}

char* gapbuf_text(gapbuf_t* g) {
    int pre_len = PRE_LEN(g);
    int post_len = POST_LEN(g);

    // Combine the pre and post segments into a single string.
    char *text = malloc(pre_len + post_len + 1);
    if (pre_len > 0) {
        memcpy(text, g->bytes, pre_len);
    }
    if (post_len > 0) {
        memcpy(text+pre_len, g->bytes+POST_START(g), post_len);
    }
    text[pre_len + post_len] = '\0';

    return text;
}

char *_gapbuf_raw_text(gapbuf_t* g) {
    int raw_text_len = g->bytes_len+1;
    char *raw_text = malloc(raw_text_len);
    memcpy(raw_text, g->bytes, g->bytes_len);
    raw_text[raw_text_len] = '\0';

    for (int i=0; i < raw_text_len-1; i++) {
        if (raw_text[i] == '\0') {
            raw_text[i] = '.';
        }
    }

    return raw_text;
}

void gapbuf_repr(gapbuf_t* g) {
    char *raw_text = _gapbuf_raw_text(g);
    char *text = gapbuf_text(g);

//    printf("bytes (%d): '%s'\n", g->bytes_len, raw_text);
//    printf("gap (%d): [%d]-[%d]\n", GAP_LEN(g), g->gap_start, g->gap_end);
    printf("text (%ld): '%s'\n", strlen(text), text);

    free(raw_text);
    free(text);
}

// Return target exponential buffer growth cap.
int _buf_cap(int len) {
    int cap = 1;
    while (cap < len) {
        cap *= 2;
    }
    return cap;
}

void _gapbuf_realloc_bytes(gapbuf_t *g, int new_bytes_len) {
    if (new_bytes_len <= g->bytes_len) return;

    // Increase capacity of buffer.
    g->bytes = realloc(g->bytes, new_bytes_len);

    // Shift post segment to the end of buffer.
    int new_gap_end = new_bytes_len - POST_LEN(g) -1;
    memcpy(g->bytes + new_gap_end + 1, g->bytes + g->gap_end + 1, POST_LEN(g));

    g->bytes_len = new_bytes_len;
    g->gap_end = new_gap_end;
}

void gapbuf_insert_text(gapbuf_t* g, char *text) {
    int text_len = strlen(text);
    if (text_len >= GAP_LEN(g)) {
        int new_bytes_len = _buf_cap(g->bytes_len + text_len);
        _gapbuf_realloc_bytes(g, new_bytes_len);
    }

    // Insert text into gap
    // Then shift gap to the right to be ready for next insert.
    memcpy(g->bytes + g->gap_start, text, text_len);
    g->gap_start += text_len;
}

void gapbuf_shift_gap(gapbuf_t *g, int shift_len) {
    if (shift_len == 0) return;

    if (shift_len < 0) {
        if (-shift_len > PRE_LEN(g)) {
            shift_len = -PRE_LEN(g);
        }

        // Shift gap to the left by shift_len bytes.
        shift_len = -shift_len;
        g->gap_start -= shift_len;
        g->gap_end -= shift_len;

        //memcpy(g->bytes + g->gap_end+1, g->bytes + g->gap_start, shift_len);
        memmove(g->bytes + g->gap_end+1, g->bytes + g->gap_start, shift_len);
        memset(g->bytes + g->gap_start, 0, GAP_LEN(g));

    } else if (shift_len > 0) {
        if (shift_len > POST_LEN(g)) {
            shift_len = POST_LEN(g);
        }

        // Shift gap to the right by shift_len bytes.
        //memcpy(g->bytes + g->gap_start, g->bytes + g->gap_end+1, shift_len);
        memmove(g->bytes + g->gap_start, g->bytes + g->gap_end+1, shift_len);

        g->gap_start += shift_len;
        g->gap_end += shift_len;

        memset(g->bytes + g->gap_start, 0, GAP_LEN(g));
    }
}

int _gapbuf_get_col(gapbuf_t *g) {
    int col = 0;
    for (int i=g->gap_start-1; i >= 0 && g->bytes[i] != '\n'; i--) {
        col++;
    }
    return col;
}

int _gapbuf_get_row(gapbuf_t *g) {
    int row = 0;
    for (int i=g->gap_start-1; i >= 0; i--) {
        if (g->bytes[i] == '\n') {
            row++;
        }
    }
    return row;
}

pos_t gapbuf_get_pos(gapbuf_t *g) {
    pos_t pos;
    pos.row = _gapbuf_get_row(g);
    pos.col = _gapbuf_get_col(g);

    return pos;
}

void _gapbuf_shift_row_back(gapbuf_t *g, int shift_rows) {
    if (shift_rows == 0) return;

    int nrows_back = 0;
    int offset = 0;
    for (int i=g->gap_start-1; i >= 0; i--) {
        offset++;

        // Move one additional row back, then one char forward to go to the
        // start of the target row
        if (g->bytes[i] == '\n') {
            nrows_back++;
            if (nrows_back == shift_rows+1) {
                offset--;
                break;
            }
        }
    }

    offset -= _gapbuf_get_col(g);
    gapbuf_shift_gap(g, -offset);
}

void _gapbuf_shift_row_forward(gapbuf_t* g, int shift_rows) {
    if (shift_rows == 0) return;

    int nrows_forward = 0;
    int offset = 0;
    for (int i=g->gap_end+1; i < g->bytes_len; i++) {
        offset++;

        if (g->bytes[i] == '\n') {
            nrows_forward++;
            if (nrows_forward == shift_rows) {
                break;
            }
        }
    }

    offset += _gapbuf_get_col(g);
    gapbuf_shift_gap(g, offset);
}

void gapbuf_shift_row(gapbuf_t* g, int shift_rows) {
    if (shift_rows < 0) {
        _gapbuf_shift_row_back(g, -shift_rows);
    } else {
        _gapbuf_shift_row_forward(g, shift_rows);
    }
}

void gapbuf_move_gap(gapbuf_t* g, pos_t dest) {
    pos_t src = gapbuf_get_pos(g);
    int delta_row = dest.row - src.row;
    int delta_col = dest.col - src.col;

    gapbuf_shift_gap(g, delta_col);
    gapbuf_shift_row(g, delta_row);
}

#ifdef TEST
int main() {
/*
    gapbuf_t* g1 = gapbuf_new();
    for (int i=0; i < 20; i++) {
        gapbuf_insert_text(g1, "abcdef ");
        gapbuf_repr(g1);
    }

    gapbuf_t* g2 = gapbuf_new();
    gapbuf_repr(g2);
    gapbuf_insert_text(g2, "0123456789");
    gapbuf_repr(g2);

    gapbuf_shift_gap(g2, -5);
    gapbuf_repr(g2);
    gapbuf_insert_text(g2, "abc");
    gapbuf_repr(g2);

    gapbuf_shift_gap(g2, -2000);
    gapbuf_insert_text(g2, "def");
    gapbuf_repr(g2);
*/

    gapbuf_t *g3 = gapbuf_new();
    gapbuf_insert_text(g3, "abcde\n12345");
    
    gapbuf_shift_gap(g3, -3);
    gapbuf_shift_row(g3, -1);
    gapbuf_repr(g3);
    //gapbuf_insert_text(g3, "*");
    //gapbuf_shift_row(g3, +1);
    //gapbuf_insert_text(g3, "$$");
    //gapbuf_repr(g3);

    return 0;
}
#endif


