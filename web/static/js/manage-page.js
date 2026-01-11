/**
 * Alpine.js component for the manage page
 * Handles CRUD operations for beans, roasters, grinders, and brewers
 */
function managePage() {
    return {
        tab: localStorage.getItem('manageTab') || 'beans',
        showBeanForm: false,
        showRoasterForm: false,
        showGrinderForm: false,
        showBrewerForm: false,
        editingBean: null,
        editingRoaster: null,
        editingGrinder: null,
        editingBrewer: null,
        beanForm: {name: '', origin: '', roast_level: '', process: '', description: '', roaster_rkey: ''},
        roasterForm: {name: '', location: '', website: ''},
        grinderForm: {name: '', grinder_type: '', burr_type: '', notes: ''},
        brewerForm: {name: '', brewer_type: '', description: ''},

        init() {
            this.$watch('tab', value => {
                localStorage.setItem('manageTab', value);
            });
            
            // Initialize cache in background
            if (window.ArabicaCache) {
                window.ArabicaCache.init();
            }
        },

        editBean(rkey, name, origin, roast_level, process, description, roaster_rkey) {
            this.editingBean = rkey;
            this.beanForm = {name, origin, roast_level, process, description, roaster_rkey: roaster_rkey || ''};
            this.showBeanForm = true;
        },

        async saveBean() {
            if (!this.beanForm.name || !this.beanForm.origin) {
                alert('Name and Origin are required');
                return;
            }

            const url = this.editingBean ? `/api/beans/${this.editingBean}` : '/api/beans';
            const method = this.editingBean ? 'PUT' : 'POST';

            const response = await fetch(url, {
                method,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(this.beanForm)
            });

            if (response.ok) {
                // Invalidate cache and reload
                if (window.ArabicaCache) {
                    window.ArabicaCache.invalidateCache();
                }
                window.location.reload();
            } else {
                const errorText = await response.text();
                alert('Failed to save bean: ' + errorText);
            }
        },

        async deleteBean(rkey) {
            if (!confirm('Are you sure you want to delete this bean?')) return;

            const response = await fetch(`/api/beans/${rkey}`, {
                method: 'DELETE',
            });
            if (response.ok) {
                // Invalidate cache and reload
                if (window.ArabicaCache) {
                    window.ArabicaCache.invalidateCache();
                }
                window.location.reload();
            } else {
                const errorText = await response.text();
                alert('Failed to delete bean: ' + errorText);
            }
        },

        editRoaster(rkey, name, location, website) {
            this.editingRoaster = rkey;
            this.roasterForm = {name, location, website};
            this.showRoasterForm = true;
        },

        async saveRoaster() {
            if (!this.roasterForm.name) {
                alert('Name is required');
                return;
            }

            const url = this.editingRoaster ? `/api/roasters/${this.editingRoaster}` : '/api/roasters';
            const method = this.editingRoaster ? 'PUT' : 'POST';

            const response = await fetch(url, {
                method,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(this.roasterForm)
            });

            if (response.ok) {
                // Invalidate cache and reload
                if (window.ArabicaCache) {
                    window.ArabicaCache.invalidateCache();
                }
                window.location.reload();
            } else {
                const errorText = await response.text();
                alert('Failed to save roaster: ' + errorText);
            }
        },

        async deleteRoaster(rkey) {
            if (!confirm('Are you sure you want to delete this roaster?')) return;

            const response = await fetch(`/api/roasters/${rkey}`, {
                method: 'DELETE',
            });
            if (response.ok) {
                // Invalidate cache and reload
                if (window.ArabicaCache) {
                    window.ArabicaCache.invalidateCache();
                }
                window.location.reload();
            } else {
                const errorText = await response.text();
                alert('Failed to delete roaster: ' + errorText);
            }
        },

        editGrinder(rkey, name, grinder_type, burr_type, notes) {
            this.editingGrinder = rkey;
            this.grinderForm = {name, grinder_type, burr_type, notes};
            this.showGrinderForm = true;
        },

        async saveGrinder() {
            if (!this.grinderForm.name || !this.grinderForm.grinder_type) {
                alert('Name and Grinder Type are required');
                return;
            }

            const url = this.editingGrinder ? `/api/grinders/${this.editingGrinder}` : '/api/grinders';
            const method = this.editingGrinder ? 'PUT' : 'POST';

            const response = await fetch(url, {
                method,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(this.grinderForm)
            });

            if (response.ok) {
                // Invalidate cache and reload
                if (window.ArabicaCache) {
                    window.ArabicaCache.invalidateCache();
                }
                window.location.reload();
            } else {
                const errorText = await response.text();
                alert('Failed to save grinder: ' + errorText);
            }
        },

        async deleteGrinder(rkey) {
            if (!confirm('Are you sure you want to delete this grinder?')) return;

            const response = await fetch(`/api/grinders/${rkey}`, {
                method: 'DELETE',
            });
            if (response.ok) {
                // Invalidate cache and reload
                if (window.ArabicaCache) {
                    window.ArabicaCache.invalidateCache();
                }
                window.location.reload();
            } else {
                const errorText = await response.text();
                alert('Failed to delete grinder: ' + errorText);
            }
        },

        editBrewer(rkey, name, brewer_type, description) {
            this.editingBrewer = rkey;
            this.brewerForm = {name, brewer_type, description};
            this.showBrewerForm = true;
        },

        editBrewerFromRow(row) {
            const rkey = row.dataset.rkey;
            const name = row.dataset.name;
            const brewer_type = row.dataset.brewerType || '';
            const description = row.dataset.description || '';
            this.editBrewer(rkey, name, brewer_type, description);
        },

        async saveBrewer() {
            if (!this.brewerForm.name) {
                alert('Name is required');
                return;
            }

            const url = this.editingBrewer ? `/api/brewers/${this.editingBrewer}` : '/api/brewers';
            const method = this.editingBrewer ? 'PUT' : 'POST';

            const response = await fetch(url, {
                method,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(this.brewerForm)
            });

            if (response.ok) {
                // Invalidate cache and reload
                if (window.ArabicaCache) {
                    window.ArabicaCache.invalidateCache();
                }
                window.location.reload();
            } else {
                const errorText = await response.text();
                alert('Failed to save brewer: ' + errorText);
            }
        },

        async deleteBrewer(rkey) {
            if (!confirm('Are you sure you want to delete this brewer?')) return;

            const response = await fetch(`/api/brewers/${rkey}`, {
                method: 'DELETE',
            });
            if (response.ok) {
                // Invalidate cache and reload
                if (window.ArabicaCache) {
                    window.ArabicaCache.invalidateCache();
                }
                window.location.reload();
            } else {
                const errorText = await response.text();
                alert('Failed to delete brewer: ' + errorText);
            }
        }
    }
}
