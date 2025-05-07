function search(o) {
    $(".search-icon").css("opacity", "1");
    var a = -1, n = 0;
    var t = typeof userDefinedSearchData !== "undefined" && userDefinedSearchData.custom ? userDefinedSearchData : {
        thisSearch: "https://www.baidu.com/s?wd=",
        thisSearchIcon: "url(" + o + ")",
        hotStatus: true,
        custom: false,
        data: [
            { name: "百度", img: "url(" + o + ") -80px 0px", position: "0px 0px", url: "https://www.baidu.com/s?wd=" },
            { name: "谷歌", img: "url(" + o + ")  -105px 0px", position: "-40px 0px", url: "https://www.google.com/search?q=" },
            { name: "必应", img: "url(" + o + ")  -80px -25px", position: "0px -40px", url: "https://cn.bing.com/search?q=" },
            { name: "好搜", img: "url(" + o + ") -105px -25px", position: "-40px -40px", url: "https://www.so.com/s?q=" },
            { name: "搜狗", img: "url(" + o + ") -80px -50px", position: "0px -80px", url: "https://www.sogou.com/web?query=" },
            { name: "淘宝", img: "url(" + o + ") -105px -50px", position: "-40px -80px", url: "https://s.taobao.com/search?q=" },
            { name: "京东", img: "url(" + o + ") -80px -75px", position: "0px -120px", url: "http://search.jd.com/Search?keyword=" },
            { name: "天猫", img: "url(" + o + ") -105px -75px", position: "-40px -120px", url: "https://list.tmall.com/search_product.htm?q=" },
            { name: "1688", img: "url(" + o + ") -80px -100px", position: "0px -160px", url: "https://s.1688.com/selloffer/offer_search.htm?keywords=" },
            { name: "知乎", img: "url(" + o + ") -105px -100px", position: "-40px -160px", url: "https://www.zhihu.com/search?type=content&q=" },
            { name: "微博", img: "url(" + o + ") -80px -125px", position: "0px -200px", url: "https://s.weibo.com/weibo/" },
            { name: "B站", img: "url(" + o + ") -105px -125px", position: "-40px -200px", url: "http://search.bilibili.com/all?keyword=" },
            { name: "豆瓣", img: "url(" + o + ") -80px -150px", position: "0px -240px", url: "https://www.douban.com/search?source=suggest&q=" },
            { name: "优酷", img: "url(" + o + ") -105px -150px", position: "-40px -240px", url: "https://so.youku.com/search_video/q_" },
            { name: "GitHub", img: "url(" + o + ") -80px -175px", position: "0px -280px", url: "https://github.com/search?utf8=✓&q=" }
        ]
    };
    var p = localStorage.getItem("searchData");
    if (p && t.custom === p.custom) {
        t = JSON.parse(p);
    }

    function u(s) {
        var i = $(s).contents().filter(function (e, r) { return r.nodeType === 3 }).text().trim();
        return i;
    }

    function l(s) {
        $.ajax({
            type: "GET",
            url: "https://sp0.baidu.com/5a1Fazu8AA54nxGko9WTAnF6hhy/su",
            async: true,
            data: { wd: s },
            dataType: "jsonp",
            jsonp: "cb",
            success: function (i) {
                $("#box ul").text("");
                n = i.s.length;
                if (n) {
                    $("#box").css("display", "block");
                    for (var e = 0; e < n; e++) {
                        $("#box ul").append("<li><span>" + (e + 1) + "</span> " + i.s[e] + "</li>");
                        $("#box ul li").eq(e).click(function () {
                            var r = u(this);
                            $("#txt").val(r);
                            window.open(t.thisSearch + r);
                            $("#box").css("display", "none");
                        });
                        if (e === 0) {
                            $("#box ul li").eq(e).css({ "border-top": "none" });
                            $("#box ul span").eq(e).css({ color: "#fff", background: "#f54545" });
                        } else if (e === 1) {
                            $("#box ul span").eq(e).css({ color: "#fff", background: "#ff8547" });
                        } else if (e === 2) {
                            $("#box ul span").eq(e).css({ color: "#fff", background: "#ffac38" });
                        }
                    }
                } else {
                    $("#box").css("display", "none");
                }
            },
            error: function (i) {
                console.log(i);
            }
        });
    }

    $("#txt").keyup(function (s) {
        if ($(this).val()) {
            if (s.keyCode == 38 || s.keyCode == 40 || !t.hotStatus) return;
            l($(this).val());
        } else {
            $(".search-clear").css("display", "none");
            $("#box").css("display", "none");
        }
    });

    $("#txt").keydown(function (s) {
        if (s.keyCode === 40) {
            a === n - 1 ? a = 0 : a++;
            $("#box ul li").eq(a).addClass("current").siblings().removeClass("current");
            var i = u($("#box ul li").eq(a));
            $("#txt").val(i);
        }
        if (s.keyCode === 38) {
            s.preventDefault && s.preventDefault();
            s.returnValue && (s.returnValue = false);
            a === 0 || a === -1 ? a = n - 1 : a--;
            $("#box ul li").eq(a).addClass("current").siblings().removeClass("current");
            var i = u($("#box ul li").eq(a));
            $("#txt").val(i);
        }
        if (s.keyCode === 13) {
            window.open(t.thisSearch + $("#txt").val());
            $("#box").css("display", "none");
            $("#txt").blur();
            $("#box ul li").removeClass("current");
            a = -1;
        }
    });

    $("#txt").focus(function () {
        $(".search-box").css("box-show", "inset 0 1px 2px rgba(27,31,35,.075), 0 0 0 0.2em rgba(3,102,214,.3)");
        $(this).val() && t.hotStatus && l($(this).val());
    });

    $("#txt").blur(function () {
        setTimeout(function () { $("#box").css("display", "none") }, 250);
    });

    for (var c = 0; c < t.data.length; c++) {
        $(".search-engine-list").append(
            '<li><span style="background:' + t.data[c].img + (t.custom ? " 0% 0% / cover no-repeat" : "") + '"/></span>' +
            t.data[c].name + "</li>"
        );
    }

    $(".search-icon, .search-engine").hover(
        function () { $(".search-engine").css("display", "block"); },
        function () { $(".search-engine").css("display", "none"); }
    );

    $("#hot-btn").click(function () {
        $(this).toggleClass("off");
        t.hotStatus = !t.hotStatus;
        localStorage.searchData = JSON.stringify(t);
    });

    t.hotStatus ? $("#hot-btn").removeClass("off") : $("#hot-btn").addClass("off");

    $(".search-engine-list li").click(function () {
        var s = $(this).index();
        t.thisSearchIcon = t.custom ? t.data[s].img : t.data[s].position;
        t.custom
            ? $(".search-icon").css("background", t.thisSearchIcon + " no-repeat").css("background-size", "cover")
            : $(".search-icon").css("background-position", t.thisSearchIcon);
        t.thisSearch = t.data[s].url;
        $(".search-engine").css("display", "none");
        localStorage.searchData = JSON.stringify(t);
    });

    t.custom
        ? $(".search-icon").css("background", t.thisSearchIcon + " no-repeat").css("background-size", "cover")
        : $(".search-icon").css("background-position", t.thisSearchIcon);

    $("#search-btn").click(function () {
        var s = $("#txt").val();
        s
            ? (window.open(t.thisSearch + s), $("#box ul").html(""))
            : layer.msg("请输入关键词！", { time: 500 }, function () { $("#txt").focus() });
    });
}

function switchNightMode() {
    var o = document.cookie.replace(/(?:(?:^|.*;\s*)night\s*\=\s*([^;]*).*$)|^.*$/, "$1") || "0";
    if (o == "0") {
        document.body.classList.add("night");
        document.cookie = "night=1;path=/";
    } else {
        document.body.classList.remove("night");
        document.cookie = "night=0;path=/";
    }
}
