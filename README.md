# gemlikes

A liking and comment system for the [Gemini](https://gemini.circumlunar.space/) protocol, especially for gemlogs! It works using CGI, so there's no additional server to run. It should work under any server that supports CGI. Tested with Jetforce, but please file a bug if it fails under a different server.

Visit `gemini://makeworld.gq/gemlog/2020-05-21-first.gmi` ([Proxy](https://portal.mozz.us/gemini/makeworld.gq/gemlog/2020-05-21-first.gmi)) to see a demo of it in action. Here's the view at the time of writing, once you click the link:

```
# 2020-05-21-first.gmi

15 likes! 💖
=> like?2020-05-21-first.gmi Add yours

12 comments 💬
=> add-comment?2020-05-21-first.gmi Add yours

wakyct (id: a1e76953) @ Mon, 25 May 2020 18:46:18 UTC:
Elphernaut checking in.

epoch (id: 2a2b33ac) @ Sun, 24 May 2020 09:51:45 UTC:
trying from castor again.

bard (id: 802ab1a9) @ Sun, 24 May 2020 05:40:47 UTC:
hello from bombadillo on guix system

u1955 (id: 44aef935) @ Sun, 24 May 2020 05:29:22 UTC:
geminawk(1) now has portable URI escaping.  (probably)

makeworld-proxy (id: a909174f) @ Sun, 24 May 2020 05:10:40 UTC:
Comment from mozz's proxy

makeworld (id: 4f9da128) @ Sun, 24 May 2020 05:02:49 UTC:
Second test from bombadillo

makeworld (id: 4f9da128) @ Sun, 24 May 2020 05:01:36 UTC:
Test from Castor

u6186 (id: 44aef935) @ Sun, 24 May 2020 04:52:19 UTC:
Hello from geminawk(1) on tilde.black.

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
There are three binaries to install: `view`, `like`, and `add-comment`.
- The binary names must not be changed
- They should be placed in a directory that will allow your server to run the binaries
  - `cgi-bin/gemlikes/` is recommended
- They all need to be in the same directory, so that relative links will work

There is also a config file that needs to be in the same directory, with the name `gemlikes.toml`. This name cannot be changed. Look at the [example-config.toml](./example-config.toml) file in the repo to see the options available. You will need to create and change the config file, it won't work without one.

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
- Filenames with spaces or special characters may cause issues for clients that don't escape their query strings. **It is recommended to only use ASCII filenames, with no spaces.** Castor and Bombadillo should be fine.
- It can't handle multiple files of the same name at different locations. For example if there is a file at `/myfile.gmi` and another file at `/dir/myfile.gmi`, gemlikes will refuse to display or perform actions, because it doesn't know which one is being referred to.
  - This only applies if both of the directories these files are in are included in the `gemlikes.toml` file. If only one is specified, such as `/dir`, there won't be any issues.
- It can't handle filenames that contain a `?` - this is to support clients that handle query strings improperly
- Comments cannot be disabled per-file, only globally. I may add this in the future.
