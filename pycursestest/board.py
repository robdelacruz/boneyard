import curses
from curses import textpad as tp

board = {
    'tiles': """
~~~.....^^
~~~....^^^
~~~.....^^
~.........
....****..
...******.
..***..***
..********
...******.
..........
    """.strip(),
    'terrain': """
www.....mm
www....mmm
www.....mm
w.........
....ffff..
...ffffff.
..fff..fff
..ffffffff
...ffffff.
..........
    """.strip()
}

ATTR_PLAINS = None
ATTR_FOREST = None
ATTR_WATER = None
ATTR_MOUNTAINS = None

cur_y,cur_x = 3,3
cur_ch = ord(' ')

def main(stdscr):
    global ATTR_PLAINS, ATTR_FOREST, ATTR_WATER, ATTR_MOUNTAINS
    global cur_y, cur_x, cur_ch

    curses.curs_set(0)      # don't show cursor
    stdscr.nodelay(True)    # don't block on getch()

    curses.init_pair(1, curses.COLOR_WHITE, curses.COLOR_BLACK)
    curses.init_pair(2, curses.COLOR_BLACK, curses.COLOR_GREEN)
    curses.init_pair(3, curses.COLOR_GREEN, curses.COLOR_BLUE)
    curses.init_pair(4, curses.COLOR_RED, curses.COLOR_BLACK)

    ATTR_PLAINS = curses.color_pair(1)
    ATTR_FOREST = curses.color_pair(2)
    ATTR_WATER = curses.color_pair(3)
    ATTR_MOUNTAINS = curses.color_pair(4)

    if not is_board_valid(board):
        raise Exception("Board is invalid.")

    stdscr.clear()
    draw_board(board, stdscr)
    draw_cursor(stdscr)

    while True:
        ch = stdscr.getch()
        if ch == -1:
            continue
        elif ch == ord('q'):
            break
        elif ch == curses.KEY_UP:
            cur_y -= 1
        elif ch == curses.KEY_DOWN:
            cur_y += 1
        elif ch == curses.KEY_LEFT:
            cur_x -= 1
        elif ch == curses.KEY_RIGHT:
            cur_x += 1
        else:
            cur_ch = ch

#        stdscr.clear()
        draw_board(board, stdscr)
        draw_cursor(stdscr)

        stdscr.refresh()


def is_board_valid(board):
    tile_lines = board['tiles'].split('\n')
    terrain_lines = board['terrain'].split('\n')

    board_height = len(tile_lines)
    board_width = len(tile_lines[0])

    if len(terrain_lines) != board_height:
        return False

    for y in range(0, board_height):
        if len(tile_lines[y]) != board_width:
            return False
        if len(terrain_lines[y]) != board_width:
            return False

    return True


def draw_board(board, w):
    tile_lines = board['tiles'].split('\n')
    terrain_lines = board['terrain'].split('\n')

    board_height = len(tile_lines)
    board_width = len(tile_lines[0])

    for y in range(0, board_height):
        for x in range(0, board_width):
            tile_ch = tile_lines[y][x]
            terrain_ch = terrain_lines[y][x]

            if tile_ch == '.':
                tile_ch = ' '

            if terrain_ch == '.':
                attr = ATTR_PLAINS
            elif terrain_ch == 'f':
                attr = ATTR_FOREST
            elif terrain_ch == 'w':
                attr = ATTR_WATER
            elif terrain_ch == 'm':
                attr = ATTR_MOUNTAINS
            else:
                attr = ATTR_PLAINS

            w.addch(y,x, ord(tile_ch), attr)


def draw_cursor(w):
    global cur_y, cur_x, cur_ch
    w.addch(cur_y, cur_x, cur_ch, curses.A_REVERSE | curses.A_BOLD)


curses.wrapper(main)

