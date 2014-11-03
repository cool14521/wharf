/*
 *  Document   : auth.js
 *  Author     : Meaglith Ma <genedna@gmail.com> @genedna
 *  Description: 
 *
 */

//Auth Page Module
angular.module('auth', ['ngRoute'])
//
.controller('SigninCtrl', ['$scope', function ($scope) {
  $scope.login = function (user) {
    console.log(user)
  // body...
  }
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
})
//directive
.directive('passwordCharactersValidator', function() {

  var REQUIRED_PATTERNS = [
    /\d+/,    //numeric values
    /[a-z]+/, //lowercase values
    /[A-Z]+/, //uppercase values
    /\W+/,    //special characters
    /^\S+$/   //no whitespace allowed
  ];

  return {
    require : 'ngModel',
    link : function($scope, element, attrs, ngModel) {
      ngModel.$validators.passwordCharacters = function(value) {
        var status = true;
        angular.forEach(REQUIRED_PATTERNS, function(pattern) {
          status = status && pattern.test(value);
        });

        return status;
      }; 
    }
  }
})
.directive('usernameAvailableValidator', ['$http', function($http) {
  return {
    require : 'ngModel',
    link : function($scope, element, attrs, ngModel) {
      ngModel.$asyncValidators.usernameAvailable = function(username) {
        return $http.get('/api/username-exists?u='+ username);
      };
    }
  }
}])
.directive('compareToValidator', function() {
  return {
    require : 'ngModel',
    link : function(scope, element, attrs, ngModel) {
      scope.$watch(attrs.compareToValidator, function() {
        ngModel.$validate();
      });

      ngModel.$validators.compareTo = function(value) {
        var other = scope.$eval(attrs.compareToValidator);
        return !value || !other || value == other;
      }
    }
  }
});
