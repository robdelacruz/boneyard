import time
import random

_pushchars = "-0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"
_last_gentime = 0
_idxs = [0] * 12

def ms_since_epoch():
  return int(time.time() * 1000)

def gen_id():
  global _pushchars, _last_gentime, _idxs

  now = ms_since_epoch()

  if now != _last_gentime:
    # Generate random indexes to pushcars
    for i in range(12):
      _idxs[i] = random.randint(0, 63)
  else:
    # Same function call time, so use incremented previous indexes
    for i in range(12):
      _idxs[i] += 1
      if _idxs[i] < 64:
        break
      _idxs[i] = 0  ## carry inc to next index pos

  _last_gentime = now

  # Prepare blank 20-char ID
  idchars = [0] * 20

  # Set cols 7 to 0 with time-indexed pushchars
  for i in range(7, -1, -1):
    idchars[i] = _pushchars[now % 64]
    now = now // 64

  # Set cols 8 to 19 with randomly indexed pushchars
  i_idxs = 0
  for i in range(8, 20):
    idx = _idxs[i_idxs]
    idchars[i] = _pushchars[idx]
    i_idxs += 1

  return ''.join(idchars)

