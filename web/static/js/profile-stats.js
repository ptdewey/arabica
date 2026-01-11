/**
 * Profile stats updater
 * Listens for HTMX content swap and updates stats from data attributes
 */

document.addEventListener('DOMContentLoaded', function() {
    // Listen for HTMX afterSwap event on the profile content
    document.body.addEventListener('htmx:afterSwap', function(evt) {
        // Only handle swaps in the profile-content element
        if (evt.detail.target.id === 'profile-content') {
            updateProfileStats();
        }
    });
});

function updateProfileStats() {
    // Get stats data from the hidden div
    const statsData = document.getElementById('profile-stats-data');
    if (!statsData) return;
    
    const stats = [
        { selector: '[data-stat="brews"]', key: 'brews' },
        { selector: '[data-stat="beans"]', key: 'beans' },
        { selector: '[data-stat="roasters"]', key: 'roasters' },
        { selector: '[data-stat="grinders"]', key: 'grinders' },
        { selector: '[data-stat="brewers"]', key: 'brewers' }
    ];
    
    stats.forEach(function(stat) {
        const value = statsData.dataset[stat.key];
        const el = document.querySelector(stat.selector);
        if (el && value !== undefined) {
            el.textContent = value;
        }
    });
}
