$(document).ready(function() {
    // 移动端侧边栏切换
    $('#sidebarToggle').on('click', function(e) {
        e.stopPropagation(); // 阻止事件冒泡
        $('.sidebar').toggleClass('show');
        $('body').toggleClass('sidebar-open');
    });

    // 点击页面任何地方关闭侧边栏（仅在移动端）
    $(document).on('click', function(e) {
        if ($(window).width() < 768 && $('.sidebar').hasClass('show')) {
            if (!$(e.target).closest('.sidebar').length && !$(e.target).closest('#sidebarToggle').length) {
                $('.sidebar').removeClass('show');
                $('body').removeClass('sidebar-open');
            }
        }
    });
    
    // 防止点击sidebar内部元素关闭sidebar
    $('.sidebar').on('click', function(e) {
        e.stopPropagation();
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