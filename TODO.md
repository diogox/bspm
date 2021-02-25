*TODO*:
* Make bspwm internal package with package for state management, and a controller package.
  * The state package starts by querying the full state, then it subscribes to events and keeps an updated internal representation of bspwm's state.
  * I'll need to add a `bspm -r` to reload the internal state by querying the full state again.
  * Then we can query bspwm through one of these packages to make bspc-go's calls more idiomatic.
  * We can then attach callbacks to certain events and use the `.On(EventType, func(){})` signature to decide what to do with them. 
    They must all be called before we start listening (so we know all the states we need to listen to. Or we could subscribe to all of them).
    This will require refactoring `bspc-go` to use only one unix socket connection to receive subscription events, otherwise 
    we can't be sure of the order they come in. Unless we use `EventTypeAll`? Not sure. The problem I was having with it is that different events 
    would sometimes arrive "joined" (missing `\n`) when they happened in quick succession.
* Write unit tests
* Improve github actions workflow? Am I using all the best build flags for a smaller binary? Move mock generation to CI and remove from repo?
  Add Makefile?
* Publish my nix package for this to the nixpkgs repo when finished.
* Finish `bspc-go` library, tag it with `v1.0.0` and update the dependency here.
* Fix a bug where, being in monocle mode, with hidden nodes, and restarting the wm (`bspc wm -r`), will cause the hidden nodes not to appear again when toggling the monocle mode off. (this is, however, fixed by running that command twice)

New Mode Ideas:
* "swallow" functionality: defined by rules sent to this tool, like `bspc rule`. And there can even be a `bspm swallow --rec` to record what node with what features (name, for eg.) is supposed to be swallowed by what node.
* "shadow". Basically you can bind a node to another, and when you run a command (or hit the hotkey that runs it), it switches between them. The use case for this is having, for example, a specific terminal window for each Goland instance. The "shadows" can never be seen, except for when you switch to it from its counterpart. We can even auto-shadows. Programs that run automatically when a condition is met in a detected new node (when I launch a Goland instance, it launches a terminal instance automatically in the background), and I can even have that instance open in the folder of the root path of whatever is open in Goland. This will require a "super-state" that all modes will need to coordinate with, to avoid having these shadows shown in "monocle-mode".