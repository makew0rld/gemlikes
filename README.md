# gemlikes

A liking and comment system for the [Gemini](https://gemini.circumlunar.space/) protocol, especially for gemlogs! It works using CGI, so there's no additional server to run.

---

**This is mostly a toy/demo project.** It works, and I run it on my gemlog, but it was quickly made and not well designed. It should be rewritten to be more robust and use a real database, but doing that is not a priority for me.

Feel free to file issues or PRs, but I won't be providing a lot of support.

---


Visit `gemini://makeworld.space/gemlog/2020-05-21-first.gmi` ([Proxy](https://portal.mozz.us/gemini/makeworld.space/gemlog/2020-05-21-first.gmi)) to see a demo of it in action. Here's an example output:

```
# 2020-05-21-first.gmi

15 likes! 💖
=> like?2020-05-21-first.gmi Add yours

4 comments 💬
=> add-comment?2020-05-21-first.gmi Add yours

ben (id: 7ec5a44d) @ Sun, 24 May 2020 04:49:11 UTC:
hello there!

makeworld (id: 4f9da128) @ Sun, 24 May 2020 04:45:06 UTC:
Test 3

makeworld (id: 4f9da128) @ Sun, 24 May 2020 04:32:26 UTC:
Test number 2

makeworld (id: 4f9da128) @ Sun, 24 May 2020 04:23:53 UTC:
Test comment.
```

Comments are displayed with the latest at the top.


## Installation

**There *was* a bug in the [Molly Brown](https://tildegit.org/solderpunk/molly-brown) Gemini server that caused gemlikes to not work.** Please update your Molly Brown to commit `2e4a10297e` or later, if you're using it. Other servers should be fine.

There are three binaries to install: `view`, `like`, and `add-comment`.
- The binary names must not be changed
- They should be placed in a directory that will allow your server to run the binaries
  - `cgi-bin/gemlikes/` is recommended
- They all need to be in the same directory, so that relative links will work

There is also a config file that needs to be in the same directory, with the name `gemlikes.toml`. This name cannot be changed. Look at the [example-config.toml](./example-config.toml) file in the repo to see the options available. You will need to create and change the config file, it won't work without one.

Finally, create a `robots.txt` file at the root of the site, and disallow any bots to access the `like` and `add-comment` binaries, to prevent accidental likes from crawlers.
Here's an example file, if the binaries are all installed in `/cgi-bin/gemlikes/`:

```robots.txt
User-agent: *
Disallow: /cgi-bin/gemlikes/like
Disallow: /cgi-bin/gemlikes/add-comment
```

## Getting binaries

The easiest option is to download the appropriate `.tar.gz` file from the releases page, extract it (`tar xvfz filename`), and move the three binaries to the right directory as outlined above.

If you have the Go toolchain installed, you can also clone the repo (Not `go get`), and then run [`single-build.sh`](./single-build.sh). The binaries will be in the newly made `build` folder, ready to be moved.

## Usage
1. Find a file you want users to be able to like (and comment) on
2. Make sure the file is in a directory that's specified in your `gemlikes.toml` file
3. Add a link to `hostname.tld/path/to/gemlikes/view?file-name.gmi`

For example, if the file is at `gemini://example.com/gemlog/first-post.gmi`, and my binaries are at `gemini://example.com/cgi-bin/gemlikes/`, here's what the file should look like:
```
<blog post text here, blah blah>

=> gemini://example.com/cgi-bin/gemlikes/view?first-post.gmi View likes and comments!
```

## Protective Measures

Gemlikes has some protections in place to prevent abuse or impersonation of the comment and liking system. Note that a server admin can make comments and likes say anything they want though.

- An ID is generated based on the commenter's IP address to prevent impersonation by other commenters
  - It's displayed right beside their username, as can be seen above
- Usernames cannot be reused on a single page by different IP addresses
- An IP address cannot make more than 5 comments on a page by default, although this is configurable in the `gemlikes.toml` file
- The same IP address cannot like a file more than one time
- Only files in the directories specified in `gemlikes.toml` can be like and commented on - Trying to reference files that don't exist will give an error

## Limitations

- It can't handle multiple files of the same name at different locations. For example if there is a file at `/myfile.gmi` and another file at `/dir/myfile.gmi`, gemlikes will refuse to display or perform actions, because it doesn't know which one is being referred to.
  - This only applies if both of the directories these files are in are included in the `gemlikes.toml` file. If only one is specified, such as `/dir`, there won't be any issues.
- Comments cannot be disabled per-file, only globally. I may add this in the future.

## License

Gemlikes is licensed under the GNU Affero General Public License, version 3. The main point of this LICENSE is that even if you modify the code and don't distribute the software to anyone, you still will have to release your changes if you use this code on a public server. Please see the [LICENSE](./LICENSE) file for details.
