// Sky background for the 3D globe view of the Places page. The texture is an equirectangular
// star map in equatorial coordinates (NASA/Goddard SVS, Deep Star Maps 2020, ESA/Gaia/DPAC).

const Deg = Math.PI / 180;

// Vertical field of view range of the sky in degrees, wider than the map's so the texture stays sharp.
export const SkyFovMin = 60;
export const SkyFovMax = 90;

// Highest number of screen pixels a texel may span before the field of view is widened.
export const MaxMagnification = 1.25;

// Opacity of the atmosphere around the globe; higher values wash out the map near its edge.
export const AtmosphereBlend = 0.3;

// Zoom level at which the atmosphere has faded out.
export const AtmosphereMaxZoom = 7;

// Style specification defaults of the sky properties that atmosphere sets.
const SkyDefaults = { "sky-color": "#88C6FC", "horizon-color": "#ffffff", "fog-color": "#ffffff", "fog-ground-blend": 0.5, "atmosphere-blend": 0.8 };

const vertexShader = `#version 300 es
in vec2 a_pos;
out vec2 v_ndc;
void main() {
  v_ndc = a_pos;
  gl_Position = vec4(a_pos, 0.0, 1.0);
}`;

const fragmentShader = `#version 300 es
precision highp float;
uniform sampler2D u_sky;
uniform vec3 u_forward;
uniform vec3 u_right;
uniform vec3 u_up;
uniform vec2 u_scale;
uniform float u_brightness;
in vec2 v_ndc;
out vec4 fragColor;
const float PI = 3.141592653589793;
void main() {
  vec3 d = normalize(u_forward + v_ndc.x * u_scale.x * u_right + v_ndc.y * u_scale.y * u_up);
  float u = 0.5 - atan(d.y, d.x) / (2.0 * PI);
  float v = 0.5 - asin(clamp(d.z, -1.0, 1.0)) / PI;
  // Takes the gradient from the longitude that is continuous here, so mipmapping leaves no seam.
  // Only the gradient is switched: switching the coordinate itself mixes values per pixel quad.
  float w = fract(u + 0.5) - 0.5;
  vec2 du = vec2(dFdx(u), dFdy(u));
  vec2 dw = vec2(dFdx(w), dFdy(w));
  vec2 dv = vec2(dFdx(v), dFdy(v));
  vec2 g = abs(dw.x) + abs(dw.y) < abs(du.x) + abs(du.y) ? dw : du;
  fragColor = vec4(textureGrad(u_sky, vec2(u, v), vec2(g.x, dv.x), vec2(g.y, dv.y)).rgb * u_brightness, 1.0);
}`;

// julianDate returns the Julian date of the specified time.
export function julianDate(date) {
  return date.getTime() / 86400000 + 2440587.5;
}

// normalizeDeg wraps an angle to the range [0, 360).
function normalizeDeg(deg) {
  return ((deg % 360) + 360) % 360;
}

// siderealTime returns the Greenwich mean sidereal time in degrees.
export function siderealTime(date) {
  return normalizeDeg(280.46061837 + 360.98564736629 * (julianDate(date) - 2451545.0));
}

// atmosphere returns a MapLibre sky that draws the atmosphere around the globe.
export function atmosphere() {
  return {
    "sky-color": "transparent",
    "horizon-color": "transparent",
    "fog-color": "transparent",
    "fog-ground-blend": 1,
    "atmosphere-blend": ["interpolate", ["linear"], ["zoom"], 0, AtmosphereBlend, AtmosphereMaxZoom - 2, AtmosphereBlend, AtmosphereMaxZoom, 0],
  };
}

// styleSky returns the sky that restores the specified style sky, since MapLibre only sets the keys passed.
export function styleSky(sky) {
  return sky && Object.keys(sky).length ? { ...SkyDefaults, ...sky } : undefined;
}

// skyScale returns the tangent of half the vertical field of view for a canvas and texture width in pixels.
export function skyScale(height, textureWidth) {
  const scale = height / ((MaxMagnification * textureWidth) / Math.PI);
  return Math.min(Math.max(scale, Math.tan((SkyFovMin * Deg) / 2)), Math.tan((SkyFovMax * Deg) / 2));
}

// cameraBasis returns the view direction and screen axes of the globe camera in equatorial coordinates.
export function cameraBasis(lng, lat, bearing, pitch, sidereal) {
  const l = (lng + sidereal) * Deg;
  const p = lat * Deg;
  const b = bearing * Deg;
  const t = pitch * Deg;

  const zenith = [Math.cos(p) * Math.cos(l), Math.cos(p) * Math.sin(l), Math.sin(p)];
  const east = [-Math.sin(l), Math.cos(l), 0];
  const north = [-Math.sin(p) * Math.cos(l), -Math.sin(p) * Math.sin(l), Math.cos(p)];

  const mix = (a, wa, c, wc) => [a[0] * wa + c[0] * wc, a[1] * wa + c[1] * wc, a[2] * wa + c[2] * wc];
  const heading = mix(north, Math.cos(b), east, Math.sin(b));

  return {
    forward: mix(zenith, -Math.cos(t), heading, Math.sin(t)),
    right: mix(east, Math.cos(b), north, -Math.sin(b)),
    up: mix(zenith, Math.sin(t), heading, Math.cos(t)),
  };
}

// compileProgram links the sky shaders and returns the program, or null on failure.
function compileProgram(gl) {
  const program = gl.createProgram();

  [
    [gl.VERTEX_SHADER, vertexShader],
    [gl.FRAGMENT_SHADER, fragmentShader],
  ].forEach(([type, source]) => {
    const shader = gl.createShader(type);
    gl.shaderSource(shader, source);
    gl.compileShader(shader);
    gl.attachShader(program, shader);
    gl.deleteShader(shader);
  });

  gl.linkProgram(program);

  if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
    gl.deleteProgram(program);
    return null;
  }

  return program;
}

// loadImage resolves with the decoded image at the specified URL.
// The CORS request lets WebGL upload an image served from a CDN.
function loadImage(url) {
  return new Promise((resolve) => {
    const img = new Image();
    img.crossOrigin = "anonymous";
    img.src = url;
    resolve(img.decode().then(() => img));
  }).catch(() => {
    throw new Error("sky: failed to load image");
  });
}

// Sky renders the star map behind a MapLibre globe, following its camera.
export class Sky {
  // constructor creates a renderer for the specified container; onLost is called if the WebGL context is lost.
  constructor(el, { brightness = 0.8, onLost = null } = {}) {
    this.el = el;
    this.brightness = brightness;
    this.onLost = onLost;
    this.map = null;
    this.canvas = null;
    this.gl = null;
    this.state = "";
    this.destroyed = false;
    this.onRender = () => this.draw();
    this.onResize = () => this.resize();
  }

  // load initializes WebGL with the star map and draws it once a map is attached.
  load(url) {
    if (this.gl) {
      return Promise.resolve();
    }

    return loadImage(url).then((img) => this.init(img));
  }

  // init uploads the decoded star map to a new WebGL2 canvas and releases the context on failure.
  init(img) {
    if (this.destroyed) {
      throw new Error("sky: destroyed");
    }

    const canvas = document.createElement("canvas");
    canvas.className = "sky-canvas";
    const gl = canvas.getContext("webgl2", { alpha: false, antialias: false, depth: false, stencil: false });

    if (!gl) {
      throw new Error("sky: WebGL2 unavailable");
    }

    try {
      if (gl.getParameter(gl.MAX_TEXTURE_SIZE) < img.naturalWidth) {
        throw new Error("sky: texture too large");
      }

      const program = compileProgram(gl);

      if (!program) {
        throw new Error("sky: failed to compile shaders");
      }

      gl.useProgram(program);

      gl.bindBuffer(gl.ARRAY_BUFFER, gl.createBuffer());
      gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 1, -1, -1, 1, 1, 1]), gl.STATIC_DRAW);
      const pos = gl.getAttribLocation(program, "a_pos");
      gl.enableVertexAttribArray(pos);
      gl.vertexAttribPointer(pos, 2, gl.FLOAT, false, 0, 0);

      gl.bindTexture(gl.TEXTURE_2D, gl.createTexture());
      gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGB8, gl.RGB, gl.UNSIGNED_BYTE, img);
      gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR);
      gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
      gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT);
      gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
      gl.generateMipmap(gl.TEXTURE_2D);

      this.uniforms = {};
      ["u_forward", "u_right", "u_up", "u_scale", "u_brightness"].forEach((name) => {
        this.uniforms[name] = gl.getUniformLocation(program, name);
      });
    } catch (err) {
      gl.getExtension("WEBGL_lose_context")?.loseContext();
      throw err;
    }

    canvas.addEventListener("webglcontextlost", () => this.lost());
    this.el.innerHTML = "";
    this.el.appendChild(canvas);
    this.canvas = canvas;
    this.textureWidth = img.naturalWidth;
    this.gl = gl;
    this.resize();
  }

  // lost destroys the renderer and notifies the owner when the browser drops its WebGL context.
  lost() {
    if (this.destroyed) {
      return;
    }

    this.destroy();
    this.onLost?.();
  }

  // attach follows the camera of the specified map.
  attach(map) {
    if (this.destroyed || this.map === map) {
      return;
    }

    this.detach();
    this.map = map;

    if (map) {
      map.on("render", this.onRender);
      map.on("resize", this.onResize);
      this.resize();
    }
  }

  // detach stops following the current map.
  detach() {
    if (this.map) {
      this.map.off("render", this.onRender);
      this.map.off("resize", this.onResize);
    }

    this.map = null;
  }

  // resize matches the canvas to its container and redraws it.
  resize() {
    if (!this.canvas) {
      return;
    }

    const ratio = Math.min(window.devicePixelRatio || 1, 2);
    const width = Math.round(this.el.clientWidth * ratio);
    const height = Math.round(this.el.clientHeight * ratio);

    if (this.canvas.width !== width || this.canvas.height !== height) {
      this.canvas.width = width;
      this.canvas.height = height;
    }

    this.state = "";
    this.draw();
  }

  // draw renders the sky when the camera of a map in globe view has changed.
  draw() {
    const gl = this.gl;
    const map = this.map;

    if (!gl || !map || gl.isContextLost() || map.getProjection?.()?.type !== "globe") {
      return;
    }

    const center = map.getCenter();
    const sidereal = siderealTime(new Date());
    const bearing = map.getBearing();
    const pitch = map.getPitch();
    const state = [center.lng, center.lat, bearing, pitch, sidereal.toFixed(2)].join();

    if (state === this.state) {
      return;
    }

    this.state = state;

    const basis = cameraBasis(center.lng, center.lat, bearing, pitch, sidereal);
    const scale = skyScale(this.canvas.height, this.textureWidth);
    const aspect = this.canvas.width / Math.max(this.canvas.height, 1);

    gl.viewport(0, 0, this.canvas.width, this.canvas.height);
    gl.uniform3fv(this.uniforms.u_forward, basis.forward);
    gl.uniform3fv(this.uniforms.u_right, basis.right);
    gl.uniform3fv(this.uniforms.u_up, basis.up);
    gl.uniform2f(this.uniforms.u_scale, scale * aspect, scale);
    gl.uniform1f(this.uniforms.u_brightness, this.brightness);
    gl.drawArrays(gl.TRIANGLE_STRIP, 0, 4);
  }

  // destroy releases the WebGL context and removes the canvas.
  destroy() {
    this.destroyed = true;
    this.detach();

    if (this.gl) {
      this.gl.getExtension("WEBGL_lose_context")?.loseContext();
    }

    if (this.canvas) {
      this.canvas.remove();
    }

    this.gl = null;
    this.canvas = null;
  }
}

// star returns a new star with random position and size.
function star(i) {
  const size = Math.round(Math.random() + 1);
  const root = document.createElement("span");

  const y = Math.floor(Math.random() * 100);
  const x = Math.floor(Math.random() * 100);
  root.style.top = y + "%";
  root.style.left = x + "%";
  root.classList.add("star", `size-${size}`, `axis-${i}`);
  return root;
}

// render draws a CSS night sky with the specified number of stars.
export function render(el, stars) {
  if (!el) {
    return;
  } else if (typeof el == "string") {
    el = document.querySelector(el);
  }

  let sky;

  if (el instanceof HTMLElement) {
    sky = el;
  } else {
    return;
  }

  if (!stars) {
    stars = 360;
  }

  sky.innerHTML = "";

  for (let i = 0; i < stars; i++) {
    sky.appendChild(star(i));
  }
}
