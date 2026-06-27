# Todo

A simple codebase comment scanner and artifact generator.

## About

A simple tool written in Go that scans your entire codebase looking for comments with various keywords such as `TODO:`, `HACK:`, `REFACTOR:` etc. and generates artifacts in various formats.

## Motivation

It's incredibly common to look through a codebase and find dozens of `/// TODO:` comments. I think it's safe to say that we're all "guilty" of it. We get into a flow state, we have deadlines, we have limited mental energy to spend hyperfixated on something that maybe could be better, but isn't necessarily worth the time and effort it takes to derail ourselves from the bigger task at hand. While most modern editors feature a todo panel or extensions to help manage them, I'm frequently in favor of using external tools for specific tasks. Because of this, I often find myself creating my own in which fit my specific needs.

For this project, I wanted a super fast tool which I could run periodically on any of my projects without needing to open the editor or run write a complicated grep command. I also wanted a tool in which I could continuously extend over time whenever I thought of something cool I could add.

## Todo..  

This project is a rewrite of my original project and is currently bare-bones and not quite ready for use. The task list 
is currently:

- [ ] Implement CLI input handling
- [ ] Split out file parsing logic and make major improvements
- [ ] Write many more tests around the existing code
- [ ] Add yaml configuration support
- [ ] Write additional output writers such as markdown, html, pdf, csv, etc.
- [ ] Work on the web api and web UI

The web portion will allow you to select a source path and then run the process, generating artifacts in various formats you select, as well as displaying the results.

## Wishlist 

1. Github issue generation / linking. 
2. Notes, comments, status, etc. options within the web UI.
3. Support for multiple languages which have different comment formats.



