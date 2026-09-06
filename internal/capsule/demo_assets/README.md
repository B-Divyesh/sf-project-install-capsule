# Hello World sample project

This bundled note describes the small public repository used by `capsule demo`.
The generated declaration installs Git and Python in a rootless Alpine capsule,
clones `octocat/Hello-World` through explicitly named HTTPS hosts, and serves
the resulting workspace only on `127.0.0.1:3000`.

The note is local sample input for review. It is never mounted into a capsule;
the install command fetches the public sample into the capsule's empty tmpfs.
