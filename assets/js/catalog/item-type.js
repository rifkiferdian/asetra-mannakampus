(function () {
    document.querySelectorAll("[data-edit-item-type]").forEach(function (button) {
        button.addEventListener("click", function () {
            var modal = document.getElementById("itemTypeEdit");
            ["id", "code", "name", "description", "active"].forEach(function (field) {
                modal.querySelector('[data-field="' + field + '"]').value = button.dataset[field] || "";
            });
            window.CatalogUI.openModal("itemTypeEdit");
        });
    });
})();
