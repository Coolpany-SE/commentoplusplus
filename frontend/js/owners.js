(function (global, document) {
  "use strict";

  (document);

  global.vueConstruct = function(callback) {
    var reactiveData = {
      owners: [],
      editingOwner: null,
    };

    global.ownersPage = new Vue({
      el: "#owners",
      data: reactiveData,
    });

    if (callback !== undefined) {
      callback();
    }
  };

  global.ownersRefresh = function() {
    var json = {
      "ownerToken": global.cookieGet("commentoOwnerToken"),
    };

    global.post(global.origin + "/api/owner/list", json, function(resp) {
      if (!resp.success) {
        global.globalErrorShow(resp.message);
        return;
      }

      for (var i in resp.owners) {
        resp.owners[i].joinDate = new Date(resp.owners[i].joinDate).toLocaleDateString();
      }

      Vue.set(global.ownersPage, "owners", resp.owners);
    });
  };

  global.settingShow = function(setting) {
    $(".pane-setting").removeClass("selected");
    $(".view").hide();
    $("#" + setting).addClass("selected");
    $("#" + setting + "-view").show();
  };

  global.ownerEditOpen = function(ownerHex) {
    var data = global.ownersPage.$data;
    
    for (var i in data.owners) {
      if (data.owners[i].ownerHex === ownerHex) {
        data.editingOwner = {
          ownerHex: data.owners[i].ownerHex,
          email: data.owners[i].email,
          name: data.owners[i].name,
          confirmedEmail: data.owners[i].confirmedEmail,
        };
        break;
      }
    }

    $("#edit-owner-email").val(data.editingOwner.email);
    $("#edit-owner-name").val(data.editingOwner.name);
    $("#edit-owner-confirmed-email").prop("checked", data.editingOwner.confirmedEmail);
    document.location.hash = "#edit-owner-modal";
  };

  global.ownerEditHandler = function() {
    var data = global.ownersPage.$data;
    
    var json = {
      "ownerToken": global.cookieGet("commentoOwnerToken"),
      "ownerHex": data.editingOwner.ownerHex,
      "email": $("#edit-owner-email").val(),
      "name": $("#edit-owner-name").val(),
      "confirmedEmail": $("#edit-owner-confirmed-email").is(":checked"),
    };

    global.buttonDisable("#save-owner-button");
    global.post(global.origin + "/api/owner/update", json, function(resp) {
      global.buttonEnable("#save-owner-button");
      
      if (!resp.success) {
        global.globalErrorShow(resp.message);
        return;
      }

      document.location.hash = "#modal-close";
      global.globalOKShow("Owner updated successfully!");
      global.ownersRefresh();
    });
  };

  global.ownerNewHandler = function() {
    var json = {
      "ownerToken": global.cookieGet("commentoOwnerToken"),
      "email": $("#new-owner-email").val(),
      "name": $("#new-owner-name").val(),
      "password": $("#new-owner-password").val(),
    };

    if (json.email === "" || json.name === "" || json.password === "") {
      global.globalErrorShow("Please fill in all fields.");
      return;
    }

    global.buttonDisable("#create-owner-button");
    global.post(global.origin + "/api/owner/new", json, function(resp) {
      global.buttonEnable("#create-owner-button");
      
      if (!resp.success) {
        global.globalErrorShow(resp.message);
        return;
      }

      document.location.hash = "#modal-close";
      $("#new-owner-email").val("");
      $("#new-owner-name").val("");
      $("#new-owner-password").val("");
      global.globalOKShow("Owner created successfully!");
      global.ownersRefresh();
    });
  };

} (window.commento, document));
