/*
 *  Document   : setting.js
 *  Author     : Meaglith Ma <genedna@gmail.com> @genedna
 *  Description:
 *
 */

'use strict';

//Auth Page Module
angular.module('setting', ['ngRoute', 'ngMessages', 'ngCookies', 'angular-growl', 'angularFileUpload'])
    .controller('SettingProfileCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', function($scope, $cookies, $http, growl, $location, $timeout, $upload) {
        var version = '1.3.8';
        $scope.fileReaderSupported = window.FileReader != null && (window.FileAPI == null || FileAPI.html5 != false);
        $scope.changeAngularVersion = function() {
            window.location.hash = $scope.angularVersion;
            window.location.reload(true);
        };
        $scope.angularVersion = window.location.hash.length > 1 ? (window.location.hash.indexOf('/') === 1 ?
            window.location.hash.substring(2) : window.location.hash.substring(1)) : '1.2.20';
        // you can also $scope.$watch('files',...) instead of calling upload from ui
        $scope.upload = function(files) {
            $scope.formUpload = false;
            if (files != null) {
                for (var i = 0; i < files.length; i++) {
                    $scope.errorMsg = null;
                    (function(file) {
                        $scope.generateThumb(file);
                        eval($scope.uploadScript);
                    })(files[i]);
                }
            }
        };
        $scope.generateThumb = function(file) {
            if (file != null) {
                if ($scope.fileReaderSupported && file.type.indexOf('image') > -1) {
                    $timeout(function() {
                        var fileReader = new FileReader();
                        fileReader.readAsDataURL(file);
                        fileReader.onload = function(e) {
                            file.dataUrl = e.target.result;
                            $timeout(function() {
                                file.upload = $upload.upload({
                                    url: '/w1/gravatar', //upload.php script, node.js route, or servlet url
                                    method: 'POST',
                                    headers: {
                                        'Content-Type': "multipart/form-data",
                                        'X-XSRFToken':base64_decode($cookies._xsrf.split('|')[0])
                                    },
                                    data: {
                                        filename: 'file'
                                    },
                                    file: file // or list of files ($files) for html5 only
                                }).progress(function(evt) {}).success(function(data, status, headers, config) { // file is uploaded successfully
                                    //console.log(data);
                                    console.log('ok');
                                }).error(function(data, status, headers, config) {
                                    console.log('err');
                                });
                                file.upload.then(function(response) {
                                    file.result = response.data;
                                }, function(response) {
                                    if (response.status > 0)
                                        $scope.errorMsg = response.status + ': ' + response.data;
                                });
                                file.upload.progress(function(evt) {
                                    file.progress = Math.min(100, parseInt(100.0 * evt.loaded / evt.total));
                                });
                            });
                        }
                    });
                }
            }
        }
        if (localStorage) {
            $scope.acl = localStorage.getItem("acl");
            $scope.success_action_redirect = localStorage.getItem("success_action_redirect");
            $scope.policy = localStorage.getItem("policy");
        }
        $scope.success_action_redirect = $scope.success_action_redirect || window.location.protocol + "//" + window.location.host;
        $scope.jsonPolicy = $scope.jsonPolicy || '{\n "expiration": "2020-01-01T00:00:00Z",\n "conditions": [\n {"bucket": "angular-file-upload"},\n ["starts-with", "$key", ""],\n {"acl": "private"},\n ["starts-with", "$Content-Type", ""],\n ["starts-with", "$filename", ""],\n ["content-length-range", 0, 524288000]\n ]\n}';
        $scope.acl = $scope.acl || 'private';

        $scope.confirm = function() {
            return confirm('Are you sure? Your local changes will be lost.');
        }
        $scope.getReqParams = function() {
            return $scope.generateErrorOnServer ? "?errorCode=" + $scope.serverErrorCode +
                "&errorMessage=" + $scope.serverErrorMsg : "";
        }
    }])
    //routes
    .config(function($routeProvider, $locationProvider) {
        $routeProvider
            .when('/', {
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