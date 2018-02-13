# acvi - acme & vi crossover

created with duit

early code. see duit.Edit documentation for how the windows work.

keyboard shortcuts combined with the command key:
- HJKL, (vi-like) to move focus to windows
- i,  to make current column wider
- I, to make current window larger
- t, to warp mouse to tag (from body)
- m, to warp mouse to body (from tag)
- s, save window
- w, close window
- n, new window
- e, execute command from header with selection in body
- 123, emulate button 1,2,3 click

extra commands in acvi:
- Open, reads the selection, interprets each line as a path and opens it.

not in acvi:
- file system interface, and the extensions it enables. i haven't used it...
- zerox, haven't used it much

## todo

- moving windows between columns

- better saving (not reading entire contents in memory and overwriting current file)
- keep track of number of lines in tag and set right height. edit will draw itself correctly on the configured height. must redo how fileui's are layed out, not using a box. not worth the trouble.

- after window close, focus on previous focussed window

- render with fixed width font
- autoindent? perhaps as part of duit.Edit.
- implement > and < for executing commands
- b1,b2,b3 drag from squares, for both files and columns.
- something like win
- structural regular expressions?
- use plumber
