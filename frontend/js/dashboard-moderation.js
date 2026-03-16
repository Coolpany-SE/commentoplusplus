(function (global, document) {
  "use strict";

  (document);

  var allCommentsControlDefaultValues = {
    allCommentsPage: 1,
    allCommentsLimit: 25,
    allCommentsSearch: "",
    allCommentsIncludeUnapproved: false,
    allCommentsIncludeDeleted: false,
  };

  global.allCommentsExtractControlsData = function(data) {
    var result = {};

    for (var key in allCommentsControlDefaultValues) {
      if (data[key] !== undefined) {
        result[key] = data[key];
      } else {
        result[key] = allCommentsControlDefaultValues[key];
      }
    }

    return result;
  }

  global.allCommentsGetInitialControlsData = function(domain) {
    var data = global.getLocalStorageData(domain + ":allCommentsControls");

    if (data == null || typeof data !== "object" || Array.isArray(data)) {
      data = {};
    }

    return global.allCommentsExtractControlsData(data);
  }

  global.allCommentsUpdateStoredControlsData = function(domain, newData) {
    var data = global.dashboard.$data;
    
    var controlsData = global.allCommentsExtractControlsData(data.domains[data.cd]);
    Object.assign(controlsData, newData);

    global.setLocalStorageData(domain + ":allCommentsControls", controlsData);
  }


  global.allCommentsWatchControls = function() {
    var data = global.dashboard.$data;
    var domain = data.domains[data.cd].domain;
    var allCommentsControlKeys = Object.keys(allCommentsControlDefaultValues);

    allCommentsControlKeys.forEach(function(key) {
      global.dashboard.$watch(
        function() {
          console.log("getter for key")
          return data.domains[data.cd][key];
        },
        function(newValue) {
          console.log("watch triggered for key", key, "newValue", newValue)
          var newData = {};
          newData[key] = newValue;
          global.allCommentsUpdateStoredControlsData(domain, newData);
        }
      );
    });
  }

  // Opens the moderatiosn settings window.
  global.moderationOpen = function() {
    $(".view").hide();
    $("#moderation-view").show();
    var data = global.dashboard.$data;
    var json = {
      "ownerToken": global.cookieGet("commentoOwnerToken"),
      "domain": data.domains[data.cd].domain
    }
    global.post(global.origin + "/api/comment/owner/list", json, function(resp) {
      if (!resp.success) {
        global.globalErrorShow(resp.message);
        return
      }
      for(var i in resp.comments) {
        resp.comments[i].creationDate = new Date(resp.comments[i].creationDate).toLocaleString()
      }
      Vue.set(data.domains[data.cd], "pending", resp.comments)
      data.domains[data.cd].pendingCommenters = resp.commenters
    });

    // Initialize all-comments controls if not already set
    if (data.domains[data.cd].allCommentsPage === undefined) {
      var initialControlsData = global.allCommentsGetInitialControlsData(data.domains[data.cd].domain);
      global.vueSet(Vue, data.domains[data.cd], initialControlsData);
      Vue.set(data.domains[data.cd], "allCommentsPagination", null);
      global.allCommentsWatchControls();
    }

    global.allCommentsRefresh();
  };

  // Fetches the "All comments" tab with current filters and pagination.
  global.allCommentsRefresh = function() {
    var data = global.dashboard.$data;
    var d = data.domains[data.cd];

    // Reset to page 1 when filters change (called from checkbox/search),
    // except when called from prev/next which set the page first.
    if (!global._allCommentsKeepPage) {
      Vue.set(d, "allCommentsPage", 1);
    }
    global._allCommentsKeepPage = false;

    var json = {
      "ownerToken": global.cookieGet("commentoOwnerToken"),
      "domain": d.domain,
      "includeUnapproved": d.allCommentsIncludeUnapproved,
      "includeDeleted": d.allCommentsIncludeDeleted,
      "page": d.allCommentsPage,
      "limit": parseInt(d.allCommentsLimit, 10)
    };

    var search = (d.allCommentsSearch || "").trim();
    if (search.length > 0) {
      json.search = search;
    }

    global.post(global.origin + "/api/comment/owner/listAll", json, function(resp) {
      if (!resp.success) {
        global.globalErrorShow(resp.message);
        return;
      }
      for (var i in resp.comments) {
        resp.comments[i].creationDate = new Date(resp.comments[i].creationDate).toLocaleString();
      }
      Vue.set(d, "commentsAll", resp.comments);
      Vue.set(d, "commentersAll", resp.commenters);
      Vue.set(d, "allCommentsPagination", resp.pagination || null);
    });
  };

  // Go to the previous page
  global.allCommentsPagePrev = function() {
    var data = global.dashboard.$data;
    var d = data.domains[data.cd];
    if (d.allCommentsPage > 1) {
      Vue.set(d, "allCommentsPage", d.allCommentsPage - 1);
      global._allCommentsKeepPage = true;
      global.allCommentsRefresh();
    }
  };

  // Go to the next page
  global.allCommentsPageNext = function() {
    var data = global.dashboard.$data;
    var d = data.domains[data.cd];
    Vue.set(d, "allCommentsPage", d.allCommentsPage + 1);
    global._allCommentsKeepPage = true;
    global.allCommentsRefresh();
  };

  // Deletes a comment from the "All comments" tab, then refreshes the list.
  global.allCommentsDeleteHandler = function(hex) {
    var json = {
      "ownerToken": global.cookieGet("commentoOwnerToken"),
      "commentHex": hex
    }
    global.post(global.origin + "/api/comment/owner/delete", json, function(resp) {
      if (!resp.success) {
        global.globalErrorShow(resp.message);
        return;
      }
      global._allCommentsKeepPage = true;
      global.allCommentsRefresh();
    });
  }

  global.moderatorRemoveComment = function(hex) {
    var data = global.dashboard.$data;
    var indexToRemove = -1;
    for (var i in data.domains[data.cd].pending) {
      if (data.domains[data.cd].pending[i].commentHex === hex) {
        indexToRemove = i;
      }
    }
    data.domains[data.cd].pending.splice(indexToRemove, 1);
  }

  // Approves a comment
  global.moderatorApproveCommentHandler = function(hex) {
    var json = {
      "ownerToken": global.cookieGet("commentoOwnerToken"),
      "commentHex": hex
    }
    global.post(global.origin + "/api/comment/owner/approve", json, function(resp) {
      if (!resp.success) {
        global.globalErrorShow(resp.message);
        return
      }
      global.moderatorRemoveComment(hex)
    })
  }


  // Deletes a comment
  global.moderatorDeleteCommentHandler = function(hex) {
    var json = {
      "ownerToken": global.cookieGet("commentoOwnerToken"),
      "commentHex": hex
    }
    global.post(global.origin + "/api/comment/owner/delete", json, function(resp) {
      if (!resp.success) {
        global.globalErrorShow(resp.message);
        return
      }
      global.moderatorRemoveComment(hex)
    })
  }

  // Adds a moderator.
  global.moderatorNewHandler = function() {
    var data = global.dashboard.$data;
    var email = $("#new-mod").val();
    
    var json = {
      "ownerToken": global.cookieGet("commentoOwnerToken"),
      "domain": data.domains[data.cd].domain,
      "email": email,
    }

    var idx = -1;
    for (var i = 0; i < data.domains[data.cd].moderators.length; i++) {
      if (data.domains[data.cd].moderators[i].email === email) {
        idx = i;
        break;
      }
    }

    if (idx === -1) {
      data.domains[data.cd].moderators.push({"email": email, "timeAgo": "just now"});
      global.buttonDisable("#new-mod-button");
      global.post(global.origin + "/api/domain/moderator/new", json, function(resp) {
        global.buttonEnable("#new-mod-button");

        if (!resp.success) {
          global.globalErrorShow(resp.message);
          return
        }

        global.globalOKShow("Added a new moderator!");
        $("#new-mod").val("");
        $("#new-mod").focus();
      });
    } else {
      global.globalErrorShow("Already a moderator.");
    }
  }


  // Deletes a moderator.
  global.moderatorDeleteHandler = function(email) {
    var data = global.dashboard.$data;
    
    var json = {
      "ownerToken": global.cookieGet("commentoOwnerToken"),
      "domain": data.domains[data.cd].domain,
      "email": email,
    }

    var idx = -1;
    for (var i = 0; i < data.domains[data.cd].moderators.length; i++) {
      if (data.domains[data.cd].moderators[i].email === email) {
        idx = i;
        break;
      }
    }

    if (idx !== -1) {
      data.domains[data.cd].moderators.splice(idx, 1);
      global.post(global.origin + "/api/domain/moderator/delete", json, function(resp) {
        if (!resp.success) {
          global.globalErrorShow(resp.message);
          return
        }

        global.globalOKShow("Removed!");
      });
    }
  }


  // Adds a domain owner.
  global.domainOwnerAddHandler = function() {
    var data = global.dashboard.$data;
    var email = $("#new-domain-owner").val();
    var addAsModerator = $("#new-domain-owner-add-moderator").is(":checked");
    
    var json = {
      "ownerToken": global.cookieGet("commentoOwnerToken"),
      "domain": data.domains[data.cd].domain,
      "email": email,
    }

    var idx = -1;
    if (data.domains[data.cd].owners) {
      for (var i = 0; i < data.domains[data.cd].owners.length; i++) {
        if (data.domains[data.cd].owners[i].email === email) {
          idx = i;
          break;
        }
      }
    }

    if (idx === -1) {
      global.buttonDisable("#new-domain-owner-button");
      global.post(global.origin + "/api/domain/owner/add", json, function(resp) {
        global.buttonEnable("#new-domain-owner-button");

        if (!resp.success) {
          global.globalErrorShow(resp.message);
          return
        }

        global.globalOKShow("Added a new domain owner!");
        $("#new-domain-owner").val("");
        $("#new-domain-owner").focus();

        // Also add as moderator if checkbox is checked
        if (addAsModerator) {
          var modJson = {
            "ownerToken": global.cookieGet("commentoOwnerToken"),
            "domain": data.domains[data.cd].domain,
            "email": email,
          };
          global.post(global.origin + "/api/domain/moderator/new", modJson, function(modResp) {
            if (!modResp.success && modResp.message !== "Already a moderator.") {
              global.globalErrorShow("Failed to add a moderator: " + modResp.message);
            }

            global.domainRefresh();
          });
        }
      });
    } else {
      global.globalErrorShow("Already a domain owner.");
    }
  }


  // Removes a domain owner.
  global.domainOwnerRemoveHandler = function(ownerHex) {
    var data = global.dashboard.$data;
    
    var json = {
      "ownerToken": global.cookieGet("commentoOwnerToken"),
      "domain": data.domains[data.cd].domain,
      "removeOwnerHex": ownerHex,
    }

    global.post(global.origin + "/api/domain/owner/remove", json, function(resp) {
      if (!resp.success) {
        global.globalErrorShow(resp.message);
        return
      }

      // Refresh the domain list to get updated owners
      global.domainRefresh(function() {
        global.globalOKShow("Removed domain owner!");
      });
    });
  }

} (window.commento, document));
