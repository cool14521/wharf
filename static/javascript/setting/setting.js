/*
 *  Document   : setting.js
 *  Author     : Meaglith Ma <genedna@gmail.com> @genedna
 *  Description:
 *
 */

'use strict';

//Auth Page Module
angular.module('setting', ['ngRoute', 'ngMessages', 'ngCookies', 'angular-growl', 'angularFileUpload'])
    .controller('SettingProfileCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', 'FileUploader', function($scope, $cookies, $http, growl, $location, $timeout, FileUploader) {
        var uploader = $scope.uploader = new FileUploader({
            url: '/'
        });

        // FILTERS

        uploader.filters.push({
            name: 'imageFilter',
            fn: function(item /*{File|FileLikeObject}*/ , options) {
                var type = '|' + item.type.slice(item.type.lastIndexOf('/') + 1) + '|';
                return '|jpg|png|jpeg|bmp|gif|'.indexOf(type) !== -1;
            }
        });
    }])
    //routes
    .config(function($routeProvider, $locationProvider) {
        $routeProvider
            .when('/', {
                templateUrl: 'static/views/setting/profile.html',
                controller: 'SettingProfileCtrl'
            })
            .when('/profile', {
                templateUrl: 'static/views/setting/profile.html',
                controller: 'SettingProfileCtrl'
            });
    })
    .directive('namespaceValidator', [function() {
        var NAMESPACE_REGEXP = /^([a-z0-9_]{6,30})$/;

        return {
            require: 'ngModel',
            restrict: '',
            link: function(scope, element, attrs, ngModel) {
                ngModel.$validators.usernames = function(value) {
                    return USERNAME_REGEXP.test(value);
                }
            }
        };
    }]);