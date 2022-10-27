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

// Since the Algolia script is loaded asynchronously, we need to wait for it to be loaded before we can use it.
document.addEventListener("DOMContentLoaded", function () {
  algoliasearchNetlify({
    appId: "{{ .Site.Params.docs_search_app_id }}",
    apiKey: "{{ .Site.Params.docs_search_api_key }}",
    siteId: "{{ .Site.Params.docs_search_site_id }}",
    branch: "main",
    selector: "div#search",
  });
  // Replace the "submit" title with a more descriptive one
  document.querySelector(".aa-SubmitButton").setAttribute("title", "Search");
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
