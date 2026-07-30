(function () {
    document.querySelectorAll("[data-edit-brand]").forEach(function (button) {
        button.addEventListener("click", function () {
            var modal = document.getElementById("brandEdit");
            ["id", "code", "name", "description", "active"].forEach(function (field) {
                modal.querySelector('[data-field="' + field + '"]').value = button.dataset[field] || "";
            });
            window.CatalogUI.openModal("brandEdit");
        });
    });
})();
