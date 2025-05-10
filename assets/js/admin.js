$(document).ready(function() {
    // 移动端侧边栏切换
    $('#sidebarToggle').on('click', function() {
        $('.sidebar').toggleClass('show');
    });

    // 点击主内容区域时关闭侧边栏（仅在移动端）
    $('.main-content').on('click', function() {
        if ($(window).width() < 768) {
            $('.sidebar').removeClass('show');
        }
    });

    // 窗口大小改变时处理侧边栏
    $(window).resize(function() {
        if ($(window).width() >= 768) {
            $('.sidebar').removeClass('show');
        }
    });

    // 表格响应式处理
    $('.table').wrap('<div class="table-responsive"></div>');

    // 添加加载动画
    $(document).ajaxStart(function() {
        $('body').append('<div class="loading"><i class="fa fa-spinner fa-spin fa-3x"></i></div>');
    }).ajaxStop(function() {
        $('.loading').remove();
    });

    // 美化提示框
    $('.alert').delay(3000).fadeOut(500);
}); 