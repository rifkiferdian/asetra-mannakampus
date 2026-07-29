(function () {
    document.querySelectorAll("[data-edit-package-item]").forEach(function (button) {
        button.addEventListener("click", function () {
            var modal = document.getElementById("packageItemEdit");
            ["id", "package", "item", "qty", "optional", "order", "notes"].forEach(function (field) {
                modal.querySelector('[data-field="' + field + '"]').value = button.dataset[field] || "";
            });
            window.CatalogUI.openModal("packageItemEdit");
        });
    });
})();
