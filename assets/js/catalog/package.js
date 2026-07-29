(function () {
    document.querySelectorAll("[data-edit-package]").forEach(function (button) {
        button.addEventListener("click", function () {
            var modal = document.getElementById("packageEdit");
            ["id", "code", "name", "category", "division", "description", "active"].forEach(function (field) {
                modal.querySelector('[data-field="' + field + '"]').value = button.dataset[field] || "";
            });
            window.CatalogUI.openModal("packageEdit");
        });
    });
})();
