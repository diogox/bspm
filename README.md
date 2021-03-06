# BSPM - The BSPWM Manager ![CI Status](https://github.com/diogox/bspm/workflows/CI/badge.svg)

`bspm` is a tool meant to patch some of `bspwm`'s shortcomings.

Here are its features:
* **Transparent Monocle Mode** - If you've used `bspwm`'s monocle mode with any sort of window transparency, 
  you've probably noticed that you can see your other open windows in the background. This is not ideal. 
  And so, `bspm` solves it! Just make sure to replace your existing hotkeys with the appropriate `bspm` commands, 
  and you're good to go!
  
* **More Coming (Hopefully) Soon!**

## Usage

First things first: To use `bspm`, you need to launch its daemon.

Simply add this to your `bspwmrc` and you'll have it up and running at startup:
```shell
bspm -d &
```

### Transparent Monocle Mode

All commands are prefixed with the subcommand `monocle`.

#### Actions
Toggle this mode for the desktop you're currently on:
```shell
bspm monocle --toggle
```

Cycle to the next node:
```shell
bspm monocle --next
```

Cycle to the previous node:
```shell
bspm monocle --prev
```

#### Subscriptions
Subscriptions are useful to create interactions with bspm's state.

You can, for example, setup your status bar to show an icon when monocle mode is active, and how many nodes are open.

Subscribe to number of nodes in the current desktop's monocle mode (number also updates as desktop focus changes):
```shell
bspm monocle --subscribe-node-count
```
*This will return `-1` if monocle mode is disabled. 
Bear in mind that a desktop in monocle mode can still have `0` nodes.*

That's it!

---

**Caution**: The `bspc node -k` command [will break this mode](https://github.com/diogox/bspm/issues/9).
You can resume it simply by toggling it off and on again.
To avoid this, you can replace that command in your `sxhkdrc` with one of the following:
* `xdo kill`
* `xdotool getwindowfocus windowkill`
* `xkill -id $(xprop -root _NET_ACTIVE_WINDOW | cut -d\# -f2)` (Not pretty, I know)

You'll need to have the necessary dependencies installed for whatever command you choose.

---

**A few tips:**: 

* To be able to use `j` and `k` to cycle between nodes in this mode and still be able to use those keys 
in `tiled` mode, you can use the following snippet in your `sxhkdrc`:
```
super + j
	bspm monocle --prev || bspc node -f south
super + k
	bspm monocle --next || bspc node -f north
```

* To get an indicator in Polybar for whether or not transparent monocle mode is activated in the currently focused desktop, 
 and how many nodes are in it, use a script like this one:
```sh
#!/bin/bash

bspm monocle --subscribe-node-count | while read -r n_nodes; do
	if [[ $n_nodes == "-1" ]]
	then 
		echo " " 
	else 
		echo "M[$n_nodes]"
	fi
done 
```
And then have a Polybar module like: (and enable it for your bar)
```
[module/bspm-monocle]
type = custom/script
exec = path_to_script.sh
tail = true
```

---

*A special thanks to [ortango](https://github.com/ortango) for his contributions in perfecting `bspm`.*