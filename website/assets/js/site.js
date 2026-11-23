(function () {
  var root = document.documentElement;
  var KEY = "zpad-theme";

  function saved() {
    try {
      return localStorage.getItem(KEY);
    } catch (e) {
      return null;
    }
  }

  function themeName() {
    return root.dataset.theme === "dark" ? "dark" : "light";
  }

  // The hero screenshot follows the theme, so a dark page never shows a light
  // window. Both files are declared in the markup and swapped here, together
  // with the lightbox link that points at the same file.
  function swapShots() {
    var theme = themeName();
    document.querySelectorAll("img[data-light][data-dark]").forEach(function (img) {
      img.src = img.getAttribute("data-" + theme) || img.getAttribute("data-light");
    });
    document.querySelectorAll("a[data-light-href]").forEach(function (a) {
      a.href = a.getAttribute("data-" + theme + "-href") || a.getAttribute("data-light-href");
    });
  }

  swapShots();

  var toggle = document.getElementById("theme-toggle");
  if (toggle) {
    toggle.addEventListener("click", function () {
      root.dataset.theme = themeName() === "dark" ? "light" : "dark";
      try {
        localStorage.setItem(KEY, root.dataset.theme);
      } catch (e) {}
      swapShots();
    });
  }

  var mq = window.matchMedia("(prefers-color-scheme: dark)");
  mq.addEventListener("change", function (event) {
    if (saved()) return;
    root.dataset.theme = event.matches ? "dark" : "light";
    swapShots();
  });

  // Screenshots open in a dialog; the links still work on their own when the
  // script does not run.
  var dialog = document.getElementById("lightbox");
  var links = Array.prototype.slice.call(document.querySelectorAll(".shot-link"));
  if (!dialog || !links.length || typeof dialog.showModal !== "function") return;

  var image = dialog.querySelector("img");
  var caption = dialog.querySelector(".lb-caption");
  var count = dialog.querySelector(".lb-count");
  var index = 0;

  function show(i) {
    index = (i + links.length) % links.length;
    var link = links[index];
    image.src = link.href;
    var figure = link.closest("figure");
    var text = "";
    if (figure) {
      var figcaption = figure.querySelector("figcaption");
      if (figcaption) text = figcaption.textContent.trim();
    }
    caption.textContent = text;
    caption.hidden = !text;
    if (count) count.textContent = (index + 1) + " / " + links.length;
  }

  links.forEach(function (link, i) {
    link.addEventListener("click", function (event) {
      if (event.metaKey || event.ctrlKey || event.shiftKey || event.button !== 0) return;
      event.preventDefault();
      show(i);
      dialog.showModal();
      dialog.focus({ preventScroll: true });
    });
  });

  dialog.addEventListener("click", function (event) {
    if (event.target === dialog) dialog.close();
  });
  dialog.querySelector('[data-lb="prev"]').addEventListener("click", function () {
    show(index - 1);
  });
  dialog.querySelector('[data-lb="next"]').addEventListener("click", function () {
    show(index + 1);
  });
  dialog.querySelector('[data-lb="close"]').addEventListener("click", function () {
    dialog.close();
  });
  dialog.addEventListener("keydown", function (event) {
    if (event.key === "ArrowLeft") {
      event.preventDefault();
      show(index - 1);
    } else if (event.key === "ArrowRight") {
      event.preventDefault();
      show(index + 1);
    }
  });
})();
