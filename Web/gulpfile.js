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
    return gulp
        .src(lessPath).pipe(watch(lessPath, function(files) {
            return files.pipe(less({
            })).pipe(gulp.dest(cssPath));
        }));
});

gulp.task('coffee', function() {
    return gulp
        .src(coffeePath).pipe(watch(coffeePath, function(files) {
            return files.pipe(coffee({
                bare: true
            }).on('error', gutil.log)).pipe(gulp.dest(jsPath));
        }));
});

gulp.task('default', ['less', 'coffee']);
