typedef struct gapbuf_t {
    char *bytes;            // total buffer storage including pre, gap, and post segments.
    int bytes_len;          // number of bytes in buffer storage.
    int gap_start, gap_end; // starting and ending indexes to gap.

    // bytes:       [0]...[bytes_len-1] 
    // gap:         [gap_start]...[gap_end]
    // pre:         [0]...[gap_start-1]
    // post:        [gap_end+1]...[bytes_len-1]
    //
    // gap_len =    gap_end-gap_start+1
    // pre_len =    gap_start
    // post_len =   bytes_len-gap_end-1 

} gapbuf_t;

typedef struct pos_t {
    int row;
    int col;
} pos_t;

gapbuf_t* gapbuf_new();
void gapbuf_free(gapbuf_t* g);
char* gapbuf_text(gapbuf_t* g);
void gapbuf_repr(gapbuf_t* g);

pos_t gapbuf_get_pos(gapbuf_t *g);
void gapbuf_insert_text(gapbuf_t* g, char *text);

void gapbuf_shift_gap(gapbuf_t *g, int shift_len);
void gapbuf_shift_row(gapbuf_t* g, int shift_rows);
void gapbuf_move_gap(gapbuf_t* g, pos_t dest);

