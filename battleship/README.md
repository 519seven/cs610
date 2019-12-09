## Prerequisites

go installed
preferably go version 1.13

## Installation

run `make`

## Launch

`./battleship -h` to view help menu
`./battleship` to access the web app on default port 5033

## Access With Browser

https://<yourserver>:5033

Login using bob, sue, or maria with passwords of B0mbs4way:(
    There is a user name elvis with password P34nutButter76

## How to play - Workflow

The workflow is as follows:

Sign up > Log in > Create a Board > Select Board > View Active Players >
Challenge Player > View Challenges > Accept Challenge > View Battle >
Enter Battle > Launch a Strike > Declare a Winner

## Most Recent Improvements

1. Removed email from user’s profile - It is PII that I don’t need for this game and I’m not going to do anything with.  I’d rather not store it.

2. Removed the ability for a user to signup with a space in their screen name.

3. Added password complexity checker (new library - GoPasswordUtilities); it uses general terms like “Horrible” and “Weak”.  If your password is too simple, I ask for mixed case, alpha-numeric and that ought to get your rating high enough.

4. Moved the player’s board to the bottom of the screen.

5. Moved the “Enter Battle” button to the top of the screen.

6. Added status to the page for the players so they know when it’s their turn.

7. Added screen tips to the checkboxes so the player can hover over and read what exactly they represent.

8. If a player attempts to take a turn before the board is “initialized”, an alert is presented to the user.

9. More improvements to Makefile - corrected the exact issues you were having since your last installation attempt.

10. Database has been changed since last time.  Added a sunkenShips counter, a winner column to track the game winner (which is getting updated upon a winning strike).

11. Revamped how I’m handling “challenger” and “opponent” on the client side of things.  I am keeping the database consistent - all challengers are challengers across all data sets; same goes for opponents; but I am now making the determination dynamically which user is which and displaying web pages accordingly.

12. Migrated away from the concept of tracking turns by userID and improved it by using a secret key; this prevents somebody from claiming to be the person whose turn it is to go all the time and winning the game via a scripted API call.  The secret is generated only after a person has had their turn.

13. More cross-checking - you can’t update your board if you’re not in a battle; you can’t update positions if you’re not in a battle; etc.

14. Added sample database that has predefined boards for users elvis and bob.


## Features Not Yet Implemented

**HIGH PRIORITY** [game flow] - Don’t permit a player to choose the same board for another battle.

**LOW PRIORITY** [usability] - Automatically refresh the list of challenges so a player who is logged in will immediately be able to select the radio button to go into their newly accepted battle.

**HIGH PRIORITY** [game flow] - Notify players of the winner when all five ships have been sunk.  This information is currently being updated in the database but still working through passing the info to the players.

**MEDIUM PRIORITY** [user information] - Current game status (“in progress”, “won”, “lost”)

**MEDIUM PRIORITY** [usability] - Once a ship has been sunk, I’d like to color it dark gray so the player knows they can move on.



## Bugs

**NORMAL** [minor annoyance] - The user is not getting a notification after a challenge has been issued.  [No work-around; When you challenge somebody, you’ll just have to trust that the person will be notified; you don’t get confirmation.]

**CRITICAL** [affects game function] - The player is permitted to use the same board for multiple games. [Work-around: Users should not select the same board; users ought to select a new board for a new challenge and make sure before you select a challenge you select a new board.]

**CRITICAL** [exposing information to user] - The player’s board name and other columns being displayed on the web pages is being shown as a full struct rather than the property that I’m after. [No work-around; Severity may be “major” but priority is low - very meaningless information is exposed to end user.]

**NORMAL** [minor annoyance; minor affect on game flow] - Polling for when a challenge has been accepted doesn’t work - this tells the challenger when their challenge has been accepted.  [Work-around: refresh page or browse away and back to “Challenges” page.]

**NORMAL** [minor annoyance; aesthetics of game board] - The alignment with the checkbox and the table cell is off just a bit on the “Opponent’s Board” displayed to the player.  [No work-around;]

**NORMAL** [minor annoyance; minor affect on game flow] - There seems to be a race condition with my routine to double-check that your board has the proper checkboxes checked.  For the dirty fix, I’m am reloading the page when this occurs which allows players to continue playing.  I estimate several hours to resolve this bug and I just found it.  Not reproducible that I’ve found yet.  [Work-around: refresh screen when the game appears to be frozen; I have put a page reload in so the end user doesn’t have to manually reload their battle screen.]

**UNKNOWN** [may have serious consequences on game flow] - Sometimes it says “You sunk their <wrong ship>” but most of the time it says “You sunk their <right ship>”.  [I think this has been fixed within the last hour so I thought I’d leave this here.]

**MINOR** - One time I was blocked from hitting “OK” on a dialog box because a tool tip popped up right when the confirmation box did.  Odd.  [Work-around: reload screen; has only happened once in all my games.]

**MAJOR** [Unable to play any games] - A newly created battle has all positions marked as “hits” for one of they players - essentially, if you’re a challenger you can only have one game going?  I only saw this for one player but I wasn’t able to get to the bottom of the issue because I just found it. [Potential for a BLOCKER if I can reproduce it...]

That said, I think I fixed it!  I was missing parenthesis around my “OR” clause in my SQL statement.  However, what I saw was only affecting one player and what I fixed ought to have fixed it for all players; it is possible that the singular player simply had an increased likelihood that it would happen to them.


### A few notes:

Encryption protects against passive attacks (eavesdropping)

Message Authentication protects against active attacks (falsification of data and transactions)

