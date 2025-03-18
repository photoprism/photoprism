// Returns a new star with random position and size.
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

// Renders a night sky with the specified number of stars.
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
