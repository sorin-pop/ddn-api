$(document).ready(function() {
  // Enable tooltips
  $(function() {
    $('[data-toggle="tooltip"]').tooltip();
    $('[data-toggle="popover"]').popover();
  });

  checkAgent();

  $("#dbname").keyup(function() {
    checkInputs();
  });

  $("#user").keyup(function() {
    checkInputs();
  });

  $("#password").keyup(function() {
    checkInputs();
  });
});

$(document).on("change", "#agent", function() {
  checkAgent();

  $("#submit").prop("disabled", false);
});

function checkInputs() {
  const userPattern = "^[a-zA-Z0-9$_]+$";
  const pwPattern = "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)[a-zA-Z\\d]{8,}$";

  var agent = $("#agent");

  agentVal = agent.val() == null ? "" : agent.val().toLowerCase();

  if (
    valid("#dbname", userPattern) &&
    valid("#user", userPattern) &&
    valid("#password", pwPattern)
  ) {
    $("button").prop("disabled", false);
  } else {
    $("button").prop("disabled", true);
  }
}

function valid(selector, pattern) {
  var value = $(selector).val();

  if (
    !(selector == "#user" && value == "root") &&
    (value.match(pattern) || value == "")
  ) {
    console.log(selector, pattern, "OK");

    $(selector)
      .parent()
      .removeClass("has-danger");
    $(selector).removeClass("form-control-danger");

    return true;
  }

  console.log(selector, pattern, "NOT OK");

  $(selector)
    .parent()
    .addClass("has-danger");
  $(selector).addClass("form-control-danger");

  return false;
}

function checkAgent() {
  var agent = $("#agent");

  if (agent && agent.length != 0) {
    $("#dbname").prop("disabled", false);
    $("#user").prop("disabled", false);
    $("#password").prop("disabled", false);
    $("#dbnamediv")
      .attr("title", "")
      .attr("data-original-title", "")
      .tooltip("hide");
    $("#userdiv")
      .attr("title", "")
      .attr("data-original-title", "")
      .tooltip("hide");
    $("#password")
      .attr("title", "")
      .attr("data-original-title", "")
      .tooltip("hide");

    agentVal = agent.val() == null ? "" : agent.val().toLowerCase();

    msg =
      "Password requires 8 alphanumerical characters with at least 1 being uppercase.";

    $("#password")
      .attr("title", msg)
      .attr("data-original-title", msg)
      .tooltip("hide");
  }

  if (agentVal.includes("oracle")) {
    msg =
      'Not needed for Oracle. Think of the User field below as the "database", as it will also be the Oracle schema that will contain the tables and their data.';

    $("#dbname").prop("disabled", true);
    $("#dbnamediv")
      .attr("data-original-title", msg)
      .tooltip("hide");
  }
}

function checkPWInput() {}

$(document).ready(function() {
  $(".table").DataTable({
    stateSave: true,
    autoWidth: false
  });
  var privateHeader = document.getElementById("private_dbs_wrapper"),
    publicHeader = document.getElementById("public_dbs_wrapper");
  if (privateHeader) {
    privateHeader.firstChild.firstChild.textContent = "Private Databases";
    privateHeader.firstChild.firstChild.className += " h3";
  }
  if (publicHeader) {
    publicHeader.firstChild.firstChild.textContent = "Public Databases";
    publicHeader.firstChild.firstChild.className += " h3";
  }
});
