$(document).ready(function () {

    setInterval(updateStatus, 2000)
    $("form[name=download]").submit(function () {
        var url = $("input[name=torrent]").val();
        if (url && url.length > 10) {
            downloadTorrent(url);
        } else {
            $("#status").html("invalid url")
        }
        return false;
    })
    $("#stopdownload").click(stopDownload)

    $("form[name=search").submit(function(){
        search($("input[name=search]").val())
        return false;
    })

    $("#playinomx").click(playInOmx)
    $("#playinplayer").click(function () {
        playInPlayer($("#playername").val())
    })

    $("#omxbackward").click(omx("backward"))
    $("#omxquit").click(omx("quit"))
    $("#omxplaypause").click(omx("playpause"))
    $("#omxforward").click(omx("forward"))
    $("#omxinfo").click(omx("info"))
    $("#omxsubs").click(omx("subs"))
    $("#omxsubsn").click(omx("subsn"))
});

function omx(cmd) {
    return function () {
        console.log("sending " + cmd + " to omx")
        $.ajax({
            type: 'POST',
            url: "/omxcmd",
            data: { cmd: cmd },
            success: function (status) {
                console.log("omxcmd:", status)
                if (status.error) {
                    $("#omxstatus").html(status.error)
                }

            }
        });
    }
}

function playInPlayer(player) {
    var url = document.location.href + "stream"
    console.log("playing " + url + " in " + player)
    $.ajax({
        type: 'POST',
        url: "/play",
        data: { url: url, player: player },
        success: function (status) {
            console.log("play in player:", status)
            if (status.error) {
                $("#omxstatus").html(status.error)
            }

        }
    });
}

function playInOmx() {
    $.ajax({
        type: 'POST',
        url: "/playinomx",
        success: function (status) {
            console.log("play in omx:", status)
            if (status.error) {
                $("#omxstatus").html(status.error)
            }
        }
    });
}

function stopDownload() {
    $.ajax({
        url: "/stopdownload"
    }).done(function (status) {
        console.log("stopdownload:", status)
        if (status.error) {
            $("#status").html(status.error)
        }
    });
}


function setVisibility(selector, visible) {
    visible ? $(selector).show() : $(selector).hide()
}

function updateStatus() {
    $.ajax({
        url: "/status"
    }).done(function (status) {
        console.log(status)

        setVisibility("#omxcontrols", status.omx_playing)
        setVisibility("#players", status.downloading && !status.omx_playing)
        setVisibility("#stopdownload", status.started)
        setVisibility("form[name=download]", !status.started)
        setVisibility(".downloadresult", !status.started)


        if (status.started) {
            var table = "<table class='pure-table pure-table-bordered'><thead><tr><td>name</td><td>progress</td><td>down</td><td>up</td><td></td></tr></thead><tr>" +
                "<td>" + status.name + "</td>" +
                "<td>" + status.progress + "</td>" +
                "<td>" + status.down + "</td>" +
                "<td>" + status.up + "</td>" +
                "<td>" + (status.started ? "&#9989;" : "") + (status.downloading ? "&#11015;" : "") + (status.omx_playing ? "&#9654;" : "") + "</td>" +
                "</tr></table>"

            $("#status").html(table)
            var streamUrl = document.location.href + "stream";
            $("#streamurl").html(" or use the stream: <a href='" + streamUrl + "' target=_blank>"+ streamUrl + "</a> somewhere elese")
        } else {
            $("#status").html("no active torrent")
        }

    });

}

function search(term) {
    $("#search").attr("disabled", "disabled")
    $.ajax({
        type: 'POST',
        url: "/search",
        data: { search: term },
        success: function (resultData) {
            var htmlTable = "no results"
            if (resultData && resultData.length) {
                htmlTable = "<table class='pure-table pure-table-bordered'><thead><tr><td>Name</td><td>Size</td><td>S/L</td><td></td></tr></thead>" +
                    resultData.map(function (result) {
                        return "<tr>" +
                            "<td>" + result.name + "</td>" +
                            "<td>" + result.size + "</td>" +
                            "<td>" + result.seed + " / " + result.leech + "</td>" +
                            "<td><button class='downloadresult' onclick='downloadTorrent(\"" + result.url + "\")'>Start Download</button></td>" +
                            "</tr>"

                    }) + "</table>"
            }
            $("#results").html(htmlTable)
            $("#search").removeAttr("disabled")
        }
    });

}

function downloadTorrent(url) {
    $.ajax({
        type: 'POST',
        url: "/download",
        data: { url: url },
        success: function (resultData) { console.log(resultData) }
    });

}