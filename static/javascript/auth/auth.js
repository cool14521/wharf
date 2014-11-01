//Auth Page Module
angular.module('auth', ['ngRoute'])
//
.controller('SigninCtrl', ['$scope', function ($scope) {
  
}])
.controller('SignupCtrl', ['$scope', function ($scope) {
  
}])
.controller('ForgotCtrl', ['$scope', function ($scope) {
  
}])
//routes
.config(function($routeProvider, $locationProvider){
  $routeProvider
    .when('/', {
      templateUrl: 'static/views/auth/signin.html',
      controller: 'SigninCtrl'
    })
    .when('/signup', {
      templateUrl: 'static/views/auth/signup.html',
      controller: 'SignupCtrl'
    })
    .when('/forgot', {
      templateUrl: 'static/views/auth/forgot.html',
      controller: 'ForgotCtrl'
    });

});
