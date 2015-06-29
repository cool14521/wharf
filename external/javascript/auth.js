/*
 *  Document   : auth.js
 *  Author     : Meaglith Ma <genedna@gmail.com> @genedna
 *  Description:
 *
 */

'use strict';

//Auth Page Module
angular.module('auth', ['ngRoute', 'ngMessages', 'ngCookies', 'angular-growl'])
    //Controllers
    .controller('SigninCtrl', ['$scope', '$cookies', '$http', 'growl', '$window', '$timeout', function($scope, $cookies, $http, growl, $window, $timeout) {

        $scope.submit = function() {
                //TODO
        }
    }])
    .controller('SignupCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', function($scope, $cookies, $http, growl, $location, $timeout) {
      
        $scope.submit = function() {
                //TODO
        }
    }])
    //routes
    .config(function($routeProvider, $locationProvider) {
        $routeProvider
            .when('/', {
                templateUrl: 'views/auth/signin.html',
                controller: 'SigninCtrl'
            })
            .when('/auth', {
                templateUrl: 'views/auth/signin.html',
                controller: 'SigninCtrl'
            })
            .when('/signup', {
                templateUrl: 'views/auth/signup.html',
                controller: 'SignupCtrl'
            });
    })
    .directive('usernameValidator', [function() {
        var USERNAME_REGEXP = /^([a-z0-9_]{6,30})$/;

        return {
            require: 'ngModel',
            restrict: '',
            link: function(scope, element, attrs, ngModel) {
                ngModel.$validators.usernames = function(value) {
                    return USERNAME_REGEXP.test(value);
                }
            }
        };

    }])
    .directive('confirmValidator', [function() {
        return {
            require: 'ngModel',
            restrict: '',
            scope: {
                passwd: "=confirmData"
            },
            link: function(scope, element, attrs, ngModel) {
                ngModel.$validators.repeat = function(value) {
                    return value == scope.passwd;
                };

                scope.$watch('passwd', function() {
                    ngModel.$validate();
                });
            }
        };
    }])
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
    .directive('passwdValidator', [function() {
        var NUMBER_REGEXP = /\d/;
        var LETTER_REGEXP = /[A-z]/;

        return {
            require: 'ngModel',
            restrict: '',
            link: function(scope, element, attrs, ngModel) {
                ngModel.$validators.passwd = function(value) {
                    return NUMBER_REGEXP.test(value) && LETTER_REGEXP.test(value);
                }
            }
        };
    }]);
