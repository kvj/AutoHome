var gulp = require('gulp');
var gutil = require('gulp-util');

var concat = require('gulp-concat');
var order = require('gulp-order');

var watch = require('gulp-watch');
var less = require('gulp-less');
var coffee = require('gulp-coffee');

var lessPath = 'static/less/**/[a-z]*.less';
var cssPath = 'static/css/';
var coffeePath = 'static/coffee/**/*.coffee';
var jsPath = 'static/js/';

gulp.task('less', function() {
    return gulp.src(lessPath).pipe(less({})).pipe(gulp.dest(cssPath));
});

gulp.task('coffee', function() {
    return gulp.src(coffeePath).pipe(coffee({bare: true}).on('error', gutil.log)).pipe(gulp.dest(jsPath));
});

gulp.task('dist', ['less', 'coffee']);
