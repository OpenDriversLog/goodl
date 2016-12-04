module.exports = function(grunt) {
    var files = {};
    files[grunt.option('out')] = grunt.option('in');
    console.log(files);
    grunt.initConfig({
        minifyPolymer: {
            default: {
                files: files
            }
        }
    });
    grunt.loadNpmTasks('grunt-minify-polymer');
};