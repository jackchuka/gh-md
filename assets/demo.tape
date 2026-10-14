# Regenerate with: vhs assets/demo.tape (from the repository root)
# Requires gh auth. Pulls only public issues from jackchuka/mdschema into a temp directory.

Output assets/demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1760
Set Height 680
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 45ms

Hide
Type `export D=/tmp/gh-md-demo GH_MD_ROOT=/tmp/gh-md && rm -rf "$D" "$GH_MD_ROOT" && mkdir -p "$D" && go build -o "$D/gh-md" . && cd "$D" && gh() { if [ "$1" = md ]; then shift; "$D/gh-md" "$@"; else command gh "$@"; fi; } && export PS1='\[\e[1;35m\]$ \[\e[0m\]' && clear`
Enter
Sleep 8s
Show

Type "gh md pull jackchuka/mdschema --issues"
Sleep 400ms
Enter
Sleep 4s

Type `gh md --list --filter 'state == "open"'`
Sleep 400ms
Enter
Sleep 3s

Type "clear"
Enter
Type "head -n 20 /tmp/gh-md/jackchuka/mdschema/issues/42.md"
Sleep 400ms
Enter
Sleep 4s

Hide
Type `perl -0pi -e 's/(<!-- gh-md:new-comment -->\n)/$1Adding a note from my local copy.\n/' /tmp/gh-md/jackchuka/mdschema/issues/42.md && clear`
Enter
Sleep 1s
Show

Type "tail -n 4 /tmp/gh-md/jackchuka/mdschema/issues/42.md"
Sleep 400ms
Enter
Sleep 2s

Type "gh md push --dry-run jackchuka/mdschema/issues/42.md"
Sleep 400ms
Enter
Sleep 5s
