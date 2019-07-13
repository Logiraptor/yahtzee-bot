# yahtzee-bot

This repository contains an implementation of the game [Yahtzee](https://en.wikipedia.org/wiki/Yahtzee), as well as four different implementations which attempt to play the game automatically.

1. Random
 - Chooses which row on the score sheet to use completely randomly; never re-rolls dice
2. Greedy
 - Chooses the row with the highest individual score; never re-rolls dice
3. Rare
 - Chooses the row which is has the lowest chance of being beaten by a later roll; never re-rolls dice
4. Greedy Mean
 - Chooses the row with the highest expected value; re-rolls dice when re-rolling has a higher expected value than the current roll

Below is the performance of each implementation across 1000 games. For each game, the same seed was used for the internal dice generator, meaning any deviation in score is due solely to the decisions made, and not due to chance.

![img](/boxplot.png)
