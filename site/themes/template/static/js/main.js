"use strict";

function mobileNavToggle() {
  const menu = document.getElementById("mobile-menu")?.parentElement;
  menu?.classList.toggle("mobile-menu-visible");
}

function docsVersionToggle() {
  const menu = document.getElementById("dropdown-menu");
  menu?.classList.toggle("dropdown-menu-visible");
}

window.onclick = function (event) {
  const target = event.target;
  const menu = document.getElementById("dropdown-menu");

  if (!target?.classList.contains("dropdown-toggle")) {
    menu?.classList.remove("dropdown-menu-visible");
  }
};

// Auto-caption tables until a custom table hook is implemented
// see https://github.com/gohugoio/hugo/issues/9316
document.querySelectorAll("table").forEach(function (table, index) {
  var caption = document.createElement("caption");
  caption.innerHTML = `Table ${index + 1}`;
  table.appendChild(caption);
});

// Remove the keyboard focus on non interactive elements
// in this case, the code blocks
// due to https://github.com/gohugoio/hugo/pull/8568,
// if adding scrollable code blocks, this can be removed/tuned-up
document.querySelectorAll("pre").forEach(function (codeBlock) {
  codeBlock.setAttribute("tabindex", "-1");
});

// Make the "cookie setting" link also clickable via keyboard for a11y purposes
// snippet based on https://community.cookiepro.com/s/article/UUID-69162cb7-c4a2-ac70-39a1-ca69c9340046
document.addEventListener("DOMContentLoaded", function () {
  var toggleDisplay = document.getElementsByClassName("ot-sdk-show-settings");
  for (var i = 0; i < toggleDisplay.length; i++) {
    toggleDisplay[i].onkeydown = function (event) {
      if (
        event.key === "Enter" ||
        event.keyCode === 13 ||
        event.key === "Space" ||
        event.keyCode === 32
      ) {
        event.stopImmediatePropagation();
        window.OneTrust.ToggleInfoDisplay();
      }
    };
  }
});
