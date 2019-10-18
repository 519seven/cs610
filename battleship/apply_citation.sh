#!/bin/bash

# execute commands to run the go program that reads
# in the first XXX lines of main.go and pastes it 
# into every other document to inform all parties 
# that this code was inspired by Let's Go by 
# Alex Edwards.  I may update the citation from 
# time to time and I don't want to have to do it
# dozens of times...propagate the message to all
# of the other files.  So, here's a script, written
# in go to take care of that tedious task for me.

GO=$(which go1.13)
$GO run apply_citation.go
if [[ $? -eq 0 ]]; then
    printf "Copied citation header to all files\n\n"
else
    printf "Error.  Citation not copied over\n\n"
fi