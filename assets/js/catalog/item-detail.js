(function () {
    document.querySelectorAll("[data-edit-detail]").forEach(function (button) {
        button.addEventListener("click", function () {
            var modal = document.getElementById("detailEdit");
            ["id", "item", "name", "value", "unit", "order"].forEach(function (field) {
                modal.querySelector('[data-field="' + field + '"]').value = button.dataset[field] || "";
            });
            window.CatalogUI.openModal("detailEdit");
        });
    });
})();
