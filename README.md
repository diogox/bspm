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

That's it!

**Here's a tip**: To be able to use `j` and `k` to cycle between nodes in this mode and still be able to use those keys 
in `tiled` mode, you can use a script like this one:

*Shell scripting isn't exactly my strong suit, (hence `bspm` being written in Go) so there might be a better way to write this script.*

```shell
#!/bin/bash
current_layout=$(bspc query -T -d | jq -r .layout)
   
if [[ $current_layout == monocle ]]
then
  if [[ $@ == up ]]
  then
    bspm monocle --next
  else
    # We assume it's "down"
    bspm monocle --prev
  fi
else
  # Act normally
  if [[ $@ == up ]]
  then
    bspc node -f north
  else
    # We assume it's "down"
    bspc node -f south
  fi
fi
```

And then have something like this in your `sxhkdrc` file:
```
super + {j,k}
  $PATH_TO_SCRIPT/script_name.sh {down,up}
```
