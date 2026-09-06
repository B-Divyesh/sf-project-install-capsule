# Demo sandbox

## Web demo

Open <https://project-install-capsule.sociobot.in/?demo=1> or use **Try it with
sample data** on the first screen. The demo loads the Alpine and Hello World
sample capsule in the browser preview.

The persistent banner says **Demo — sample data, nothing is saved**. **Reset
demo** restores the shipped values. **Start for real** removes the `demo:`
browser-storage entry and returns to the normal empty preview. Normal preview
edits do not use browser storage.

Demo state uses only the `localStorage` key `demo:capsule-composer`. It never
reads or writes a normal-storage key.

## CLI demo

Run:

```sh
capsule demo
```

The command creates a new `capsule-demo-*` folder under the system temporary
directory, writes `capsule.json` and `sample-project/README.md`, and prints the
review. It does not start a container. To keep the files, give it a path that
does not already exist:

```sh
capsule demo --dir ./capsule-sample
```

The demo is reset by running it again in a different new folder. The bundled
source sample is [examples/hello-world](../examples/hello-world/).
