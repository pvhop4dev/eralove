#!/usr/bin/env node

/**
 * Newman Test Runner for EraLove API
 * 
 * This script runs automated tests using Newman (Postman CLI)
 * Usage: node newman-test-script.js [environment]
 */

const newman = require('newman');
const path = require('path');
const fs = require('fs');

// Configuration
const config = {
    collection: path.join(__dirname, 'EraLove_API_Collection.json'),
    environments: {
        dev: path.join(__dirname, 'EraLove_Development_Environment.json'),
        prod: path.join(__dirname, 'EraLove_Production_Environment.json')
    },
    reporters: ['cli', 'html', 'json'],
    outputDir: path.join(__dirname, 'test-results'),
    timeout: 30000,
    iterations: 1
};

// Ensure output directory exists
if (!fs.existsSync(config.outputDir)) {
    fs.mkdirSync(config.outputDir, { recursive: true });
}

// Get environment from command line argument
const environment = process.argv[2] || 'dev';
const environmentFile = config.environments[environment];

if (!environmentFile || !fs.existsSync(environmentFile)) {
    console.error(`❌ Environment file not found: ${environment}`);
    console.log('Available environments: dev, prod');
    process.exit(1);
}

console.log(`🚀 Starting EraLove API Tests`);
console.log(`📁 Collection: ${config.collection}`);
console.log(`🌍 Environment: ${environment}`);
console.log(`📊 Output Directory: ${config.outputDir}`);
console.log('─'.repeat(50));

// Newman run configuration
const runOptions = {
    collection: config.collection,
    environment: environmentFile,
    reporters: config.reporters,
    reporter: {
        html: {
            export: path.join(config.outputDir, `eralove-test-report-${environment}-${Date.now()}.html`)
        },
        json: {
            export: path.join(config.outputDir, `eralove-test-results-${environment}-${Date.now()}.json`)
        }
    },
    timeout: config.timeout,
    iterationCount: config.iterations,
    bail: false, // Continue on failures
    color: 'on',
    verbose: true
};

// Run Newman
newman.run(runOptions, function (err, summary) {
    if (err) {
        console.error('❌ Newman run failed:', err);
        process.exit(1);
    }

    console.log('\n' + '='.repeat(50));
    console.log('📊 TEST SUMMARY');
    console.log('='.repeat(50));

    // Print summary statistics
    const stats = summary.run.stats;
    const failures = summary.run.failures;

    console.log(`📋 Total Requests: ${stats.requests.total}`);
    console.log(`✅ Passed: ${stats.requests.total - stats.requests.failed}`);
    console.log(`❌ Failed: ${stats.requests.failed}`);
    console.log(`⏱️  Average Response Time: ${Math.round(stats.requests.average)}ms`);
    console.log(`🕐 Total Test Duration: ${Math.round(summary.run.timings.completed / 1000)}s`);

    // Print test results
    if (stats.tests) {
        console.log(`\n🧪 Test Assertions:`);
        console.log(`   ✅ Passed: ${stats.tests.total - stats.tests.failed}`);
        console.log(`   ❌ Failed: ${stats.tests.failed}`);
    }

    // Print failures if any
    if (failures && failures.length > 0) {
        console.log('\n❌ FAILURES:');
        console.log('─'.repeat(30));
        failures.forEach((failure, index) => {
            console.log(`${index + 1}. ${failure.source.name || 'Unknown Request'}`);
            console.log(`   Error: ${failure.error.message}`);
            if (failure.error.test) {
                console.log(`   Test: ${failure.error.test}`);
            }
            console.log('');
        });
    }

    // Print report locations
    console.log('\n📄 Reports Generated:');
    console.log(`   HTML: ${runOptions.reporter.html.export}`);
    console.log(`   JSON: ${runOptions.reporter.json.export}`);

    // Exit with appropriate code
    const exitCode = stats.requests.failed > 0 ? 1 : 0;
    console.log(`\n${exitCode === 0 ? '✅ All tests passed!' : '❌ Some tests failed!'}`);
    process.exit(exitCode);
});

// Handle process termination
process.on('SIGINT', () => {
    console.log('\n⚠️  Test run interrupted by user');
    process.exit(1);
});

process.on('SIGTERM', () => {
    console.log('\n⚠️  Test run terminated');
    process.exit(1);
});
