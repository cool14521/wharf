/*
 *  Document   : header.js
 *  Author     : Chen Liang <allen@docker.cn.> @chliang2030598
 *  Description:
 *
 */

'use strict';
angular.module('Header', ['ngRoute', 'ngMessages', 'ngCookies', 'angular-growl', 'angularFileUpload'])
    .controller('HeaderController', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
        $scope.aaa = "aaaaa";
        
    }]);