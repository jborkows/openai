if (!window.errorHandler) {
  init();
}

function init() {
  const dialog = document.getElementById("errorDialog");
  const content = document.getElementById("errorDialogContent");
  dialog.addEventListener("click", function (evt) {
    if (!content.contains(evt.target)) {
      dialog.close();
    }
  });
  window.errorHandler = function (evt) {
    content.innerHTML = evt.detail.error;
    dialog.showModal();
  };
  document.body.addEventListener("htmx:responseError", window.errorHandler);
}
