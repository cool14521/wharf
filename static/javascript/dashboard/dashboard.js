/*
 *  Document   : setting.js
 *  Author     : Meaglith Ma <genedna@gmail.com> @genedna
 *  Description:
 *
 */

'use strict';

//Auth Page Module
angular.module('dashboard', ['ngRoute', 'ngMessages', 'ngCookies', 'angular-growl', 'ui.codemirror'])
    .controller('AddRepositoryCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $window) {
        $scope.privated = {};

        //init user data
        $scope.addPrivilege = false
        $http.get('/w1/namespace')
            .success(function(data, status, headers, config) {
                $scope.namespaces = data
            })
            .error(function(data, status, headers, config) {

            });
        $scope.privated = {};
        $scope.privated.values = [{
            code: 0,
            name: "Public"
        }, {
            code: 1,
            name: "Private"
        }];
        $scope.privated.selection = $scope.privated.values[0];

        //deal with create repository
        $scope.createRepo = function() {
            if ($scope.privated.selection.code == 1) {
                $scope.repository.privated = true;
            } else {
                $scope.repository.privated = false;
            }

            $http.defaults.headers.post['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
            $http.post('/w1/repository', $scope.repository)
                .success(function(data, status, headers, config) {
                    $scope.addPrivilege = true ;
                    growl.info(data.message);
                })
                .error(function(data, status, headers, config) {
                    growl.error(data.message);
                });
        }

    }])
    .controller('PublicRepositoryCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $window) {

    }])
    .controller('RepositoriesCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $window) {

    }])
    .controller('PrivateRepositoryCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $window) {

    }])
    .controller('StarRepositoryCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $window) {

    }])
    .controller('DockerfileCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $window) {

    }])
    //routes
    .config(function($routeProvider, $locationProvider) {
        $routeProvider
            .when('/', {
                templateUrl: '/static/views/dashboard/repositories.html',
                controller: 'RepositoriesCtrl'
            })
            .when('/repo', {
                templateUrl: '/static/views/dashboard/repositories.html',
                controller: 'RepositoriesCtrl'
            })
            .when('/repo/add', {
                templateUrl: '/static/views/dashboard/repositoryadd.html',
                controller: 'AddRepositoryCtrl'
            })
            .when('/repo/public', {
                templateUrl: '/static/views/dashboard/repopublic.html',
                controller: 'PublicRepositoryCtrl'
            })
            .when('/repo/private', {
                templateUrl: '/static/views/dashboard/repoprivate.html',
                controller: 'PrivateRepositoryCtrl'
            })
            .when('/repo/star', {
                templateUrl: '/static/views/dashboard/repostar.html',
                controller: 'StarRepositoryCtrl'
            })
            .when('/comments', {
                templateUrl: '/static/views/dashboard/comment.html',
                controller: 'CommentCtrl'
            })
            .when('/repo/dockerfile', {
                templateUrl: '/static/views/dashboard/dockerfile.html',
                controller: 'DockerfileCtrl'
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