# GPUI
A small binary to display (and kinda control) the current playing media on linux

## Running
1. Clone this Repository
2. Build the project from source 
3. Run it in the terminal with or without media playing in the background
4. Enjoy the view

## External Dependencies
It uses `playerctl` under the hood to abstract the different media players.
The End User also needs to have this program installed

## Additional Info 
Just learning go and just needed a quick terminal ui music player 
aka. the bad code was foreseeable. 

When pausing the current time still traverses because 
playerctl is kinda garbage and I am too lazy to deal with that error rn
