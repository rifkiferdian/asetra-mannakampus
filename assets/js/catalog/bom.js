(function () {
    document.querySelectorAll("[data-edit-bom]").forEach(function (button) {
        button.addEventListener("click", function () {
            var modal = document.getElementById("bomEdit");
            ["id", "parent", "code", "name", "version", "description", "active"].forEach(function (field) {
                modal.querySelector('[data-field="' + field + '"]').value = button.dataset[field] || "";
            });
            window.CatalogUI.openModal("bomEdit");
        });
    });
})();
