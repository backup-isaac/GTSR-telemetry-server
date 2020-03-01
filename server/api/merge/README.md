# WTF Are All These Static Files?

When you visit the `/merge` URL in a browser (ex. `solarracing.me/merge` or `localhost:8888/merge`), you're served a React application. This small web app is developed in the `telemetry-server-merge-webpage` repo. A production build of that project produces the following static files and directories:

* `index.html`
* `asset-manifest.json`
* `precache-manifest.3b68b775eb4c25cc37bf1ffd2517c3c8.js`
* `service-worker.js`
* `/static`, and all of its contents

# Do They All Have To Be Here?

For now: ya probably. I haven't found a better way to organize this stuff as of now.
