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
        //init user info
        $http.get('/w1/profile')
            .success(function(data, status, headers, config) {
                $scope.user = data
            })
            .error(function(data, status, headers, config) {

                $timeout(function() {
                    $window.location.href = '/auth';
                }, 3000);
            });

        //deal with fileupload start
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
                                        'X-XSRFToken': base64_decode($cookies._xsrf.split('|')[0])
                                    },
                                    data: {
                                        filename: 'file'
                                    },
                                    file: file // or list of files ($files) for html5 only
                                }).progress(function(evt) {
                                    file.progress = Math.min(100, parseInt(100.0 * evt.loaded / evt.total));
                                }).success(function(data, status, headers, config) { // file is uploaded successfully
                                    growl.info(data.message);
                                    $scope.user.gravatar=data.url
                                }).error(function(data, status, headers, config) {
                                    growl.error(data.message);
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
            //deal with fileupload end
        $scope.submit = function() {
            if ($scope.profileForm.$valid) {
                $http.defaults.headers.put['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
                $http.put('/w1/profile', $scope.user)
                    .success(function(data, status, headers, config) {
                        growl.info(data.message);
                        $timeout(function() {

                        }, 1000);
                    })
                    .error(function(data, status, headers, config) {
                        $scope.submitting = false;
                        growl.error(data.message);
                        $timeout(function() {

                        }, 1000);
                    });
            }
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
    .directive('emailValidator', [function() {
        var EMAIL_REGEXP = /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;

        return {
            require: 'ngModel',
            restrict: '',
            link: function(scope, element, attrs, ngModel) {
                ngModel.$validators.emails = function(value) {
                    return EMAIL_REGEXP.test(value);
                }
            }
        };
    }])
    .directive('urlValidator', [function() {
        var URL_REGEXP = /(http|https):\/\/[\w\-_]+(\.[\w\-_]+)+([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?/;

        return {
            require: 'ngModel',
            restrict: '',
            link: function(scope, element, attrs, ngModel) {
                ngModel.$validators.urls = function(value) {
                    return URL_REGEXP.test(value);
                }
            }
        };
    }])
    .directive('mobileValidator', [function() {
        var MOBILE_REGEXP = /^[0-9]{1,20}$/;

        return {
            require: 'ngModel',
            restrict: '',
            link: function(scope, element, attrs, ngModel) {
                ngModel.$validators.mobiles = function(value) {
                    return MOBILE_REGEXP.test(value);
                }
            }
        };
    }])
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